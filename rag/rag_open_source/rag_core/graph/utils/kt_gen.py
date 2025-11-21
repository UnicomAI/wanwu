import json
import os
import threading
import time
from concurrent import futures
from typing import Any, Dict, List, Tuple

import nanoid
import networkx as nx
import tiktoken
import json_repair

from config import get_config
from utils import call_llm_api, graph_processor
from utils.logger import logger


class KTBuilder:
    def __init__(self, dataset_name, embedding_model, schema_path=None, mode=None, config=None, schema=None):
        if config is None:
            config = get_config()

        self.config = config
        self.dataset_name = dataset_name
        if schema:
            self.schema = schema
        else:
            self.schema = self.load_schema(schema_path or config.get_dataset_config(dataset_name).schema_path)
        self.graph = nx.MultiDiGraph()
        self.node_counter = 0
        self.datasets_no_chunk = config.construction.datasets_no_chunk
        self.token_len = 0
        self.lock = threading.Lock()
        self.llm_client = call_llm_api.LLMCompletionCall(config.construction.LLM_MODEL,
                                                         config.construction.LLM_BASE_URL,
                                                         config.construction.LLM_API_KEY,
                                                         config.construction.TEMPERATURE)
        self.all_chunks = {}
        self.embedding_model = embedding_model
        self.mode = mode or config.construction.mode

    def load_schema(self, schema_path) -> Dict[str, Any]:
        try:
            with open(schema_path) as f:
                schema = json.load(f)
                return schema
        except FileNotFoundError:
            return dict()

    def chunk_text(self, text) -> Tuple[List[str], Dict[str, str]]:
        # if self.dataset_name in self.datasets_no_chunk:
        #     chunks = [f"{text.get('title', '')} {text.get('snippet', '')} {text['meta_data']['chunk_current_num']}".strip()
        #              if isinstance(text, dict) else str(text)]
        # else:
        #     chunks = [str(text)]
        # ===== Save complete information directly =====
        chunks = [text]
        chunk2id = {}
        for chunk in chunks:
            try:
                chunk_id = nanoid.generate(size=8)
                chunk2id[chunk_id] = chunk
            except Exception as e:
                logger.warning(f"Failed to generate chunk id with nanoid: {type(e).__name__}: {e}")

        with self.lock:
            self.all_chunks.update(chunk2id)

        return chunks, chunk2id

    def _clean_text(self, text: str) -> str:
        if not text:
            return "[EMPTY_TEXT]"

        if self.dataset_name == "graphrag-bench":
            safe_chars = {
                *" .:,!?()-+=[]{}()\\/|_^~<>*&%$#@!;\"'`"
            }
            cleaned = "".join(
                char for char in text
                if char.isalnum() or char.isspace() or char in safe_chars
            ).strip()
        else:
            safe_chars = {
                *" .:,!?()-+="
            }
            cleaned = "".join(
                char for char in text
                if char.isalnum() or char.isspace() or char in safe_chars
            ).strip()

        return cleaned if cleaned else "[EMPTY_AFTER_CLEANING]"

    def save_chunks_to_file(self):
        os.makedirs("output/chunks", exist_ok=True)
        chunk_file = f"output/chunks/{self.dataset_name}.txt"

        existing_data = {}
        if os.path.exists(chunk_file):
            try:
                with open(chunk_file, "r", encoding="utf-8") as f:
                    for line in f:
                        line = line.strip()
                        if line and "\t" in line:
                            # Parse line format: "id: {id} \tChunk: {chunk text}"
                            parts = line.split("\t", 1)
                            if len(parts) == 2 and parts[0].startswith("id: ") and parts[1].startswith("Chunk: "):
                                chunk_id = parts[0][4:]
                                chunk_text = parts[1][7:]
                                existing_data[chunk_id] = chunk_text
            except Exception as e:
                logger.warning(f"Failed to parse existing chunks from {chunk_file}: {type(e).__name__}: {e}")

        all_data = {**existing_data, **self.all_chunks}

        with open(chunk_file, "w", encoding="utf-8") as f:
            for chunk_id, chunk_text in all_data.items():
                f.write(f"id: {chunk_id}\tChunk: {chunk_text}\n")

        logger.info(f"Chunk data saved to {chunk_file} ({len(all_data)} chunks)")

    def extract_with_llm(self, prompt: str):
        response = self.llm_client.call_api(prompt)
        parsed_dict = json_repair.loads(response)
        parsed_json = json.dumps(parsed_dict, ensure_ascii=False)
        return parsed_json

    def token_cal(self, text: str):
        encoding = tiktoken.get_encoding("cl100k_base")
        return len(encoding.encode(text))

    def _get_construction_prompt(self, chunk: str) -> str:
        """Get the appropriate construction prompt based on dataset name."""
        recommend_schema = json.dumps(self.schema, ensure_ascii=False)

        prompt_type_map = {
            "novel": "novel_chs",
            "novel_eng": "novel_eng"
        }

        prompt_type = prompt_type_map.get(self.dataset_name, "general")
        chunk_str = f"{chunk['title']} {chunk['snippet']}".strip() if isinstance(chunk, dict) else str(chunk)
        return self.config.get_prompt_formatted("construction", prompt_type, schema=recommend_schema, chunk=chunk_str)

    def _validate_and_parse_llm_response(self, prompt: str, llm_response: str) -> dict:
        """Validate and parse LLM response, returning None if invalid."""
        if llm_response is None:
            return None

        try:
            self.token_len += self.token_cal(prompt + llm_response)
            return json_repair.loads(llm_response)
        except Exception as e:
            llm_response_str = str(llm_response) if llm_response is not None else "None"
            return None

    def _find_or_create_entity(self, entity_name: str, chunk_id: int, nodes_to_add: list,
                               entity_type: str = None) -> str:
        """Find existing entity or create a new one, returning the entity node ID."""
        with self.lock:
            entity_node_id = next(
                (
                    n
                    for n, d in self.graph.nodes(data=True)
                    if d.get("label") == "entity" and d["properties"]["name"] == entity_name
                ),
                None,
            )

            if not entity_node_id:
                entity_node_id = f"entity_{self.node_counter}"
                properties = {"name": entity_name, "chunk id": chunk_id}
                if entity_type:
                    properties["schema_type"] = entity_type

                nodes_to_add.append((
                    entity_node_id,
                    {
                        "label": "entity",
                        "properties": properties,
                        "level": 2
                    }
                ))
                self.node_counter += 1

        return entity_node_id

    def _validate_triple_format(self, triple: list) -> tuple:
        """Validate and normalize triple format, returning (subject, predicate, object) or None."""
        try:
            if len(triple) > 3:
                triple = triple[:3]
            elif len(triple) < 3:
                return None

            return tuple(triple)
        except Exception as e:
            return None

    def _process_attributes(self, extracted_attr: dict, chunk_id: int, entity_types: dict = None) -> tuple[list, list]:
        """Process extracted attributes and return nodes and edges to add."""
        nodes_to_add = []
        edges_to_add = []

        for entity, attributes in extracted_attr.items():
            for attr in attributes:
                # Create attribute node
                attr_node_id = f"attr_{self.node_counter}"
                nodes_to_add.append((
                    attr_node_id,
                    {
                        "label": "attribute",
                        "properties": {"name": attr, "chunk id": chunk_id},
                        "level": 1,
                    }
                ))
                self.node_counter += 1

                entity_type = entity_types.get(entity) if entity_types else None
                entity_node_id = self._find_or_create_entity(entity, chunk_id, nodes_to_add, entity_type)
                edges_to_add.append((entity_node_id, attr_node_id, "has_attribute"))

        return nodes_to_add, edges_to_add

    def _process_triples(self, extracted_triples: list, chunk_id: int, entity_types: dict = None) -> tuple[list, list]:
        """Process extracted triples and return nodes and edges to add."""
        nodes_to_add = []
        edges_to_add = []

        for triple in extracted_triples:
            validated_triple = self._validate_triple_format(triple)
            if not validated_triple:
                continue

            subj, pred, obj = validated_triple

            subj_type = entity_types.get(subj) if entity_types else None
            obj_type = entity_types.get(obj) if entity_types else None

            subj_node_id = self._find_or_create_entity(subj, chunk_id, nodes_to_add, subj_type)
            obj_node_id = self._find_or_create_entity(obj, chunk_id, nodes_to_add, obj_type)

            edges_to_add.append((subj_node_id, obj_node_id, pred))

        return nodes_to_add, edges_to_add

    def process_level1_level2(self, chunk: str, id: int):
        """Process attributes (level 1) and triples (level 2) with optimized structure."""
        prompt = self._get_construction_prompt(chunk)
        llm_response = self.extract_with_llm(prompt)

        # Validate and parse response
        parsed_response = self._validate_and_parse_llm_response(prompt, llm_response)
        if not parsed_response:
            return

        extracted_attr = parsed_response.get("attributes", {})
        extracted_triples = parsed_response.get("triples", [])
        entity_types = parsed_response.get("entity_types", {})

        # Process attributes and triples
        attr_nodes, attr_edges = self._process_attributes(extracted_attr, id, entity_types)
        triple_nodes, triple_edges = self._process_triples(extracted_triples, id, entity_types)

        all_nodes = attr_nodes + triple_nodes
        all_edges = attr_edges + triple_edges

        with self.lock:
            for node_id, node_data in all_nodes:
                self.graph.add_node(node_id, **node_data)

            for u, v, relation in all_edges:
                self.graph.add_edge(u, v, relation=relation)

    def _find_or_create_entity_direct(self, entity_name: str, chunk_id: int, entity_type: str = None) -> str:
        """Find existing entity or create a new one directly in graph (for agent mode)."""
        entity_node_id = next(
            (
                n
                for n, d in self.graph.nodes(data=True)
                if d.get("label") == "entity" and d["properties"]["name"] == entity_name
            ),
            None,
        )

        if not entity_node_id:
            entity_node_id = f"entity_{self.node_counter}"
            properties = {"name": entity_name, "chunk id": chunk_id}
            if entity_type:
                properties["schema_type"] = entity_type

            self.graph.add_node(
                entity_node_id,
                label="entity",
                properties=properties,
                level=2
            )
            self.node_counter += 1

        return entity_node_id

    def _process_attributes_agent(self, extracted_attr: dict, chunk_id: int, entity_types: dict = None):
        """Process extracted attributes in agent mode (direct graph operations)."""
        for entity, attributes in extracted_attr.items():
            for attr in attributes:
                # Create attribute node
                attr_node_id = f"attr_{self.node_counter}"
                self.graph.add_node(
                    attr_node_id,
                    label="attribute",
                    properties={
                        "name": attr,
                        "chunk id": chunk_id
                    },
                    level=1,
                )
                self.node_counter += 1

                entity_type = entity_types.get(entity) if entity_types else None
                entity_node_id = self._find_or_create_entity_direct(entity, chunk_id, entity_type)
                self.graph.add_edge(entity_node_id, attr_node_id, relation="has_attribute")

    def _process_triples_agent(self, extracted_triples: list, chunk_id: int, entity_types: dict = None):
        """Process extracted triples in agent mode (direct graph operations)."""
        for triple in extracted_triples:
            validated_triple = self._validate_triple_format(triple)
            if not validated_triple:
                continue

            subj, pred, obj = validated_triple

            subj_type = entity_types.get(subj) if entity_types else None
            obj_type = entity_types.get(obj) if entity_types else None

            # Find or create subject and object entities
            subj_node_id = self._find_or_create_entity_direct(subj, chunk_id, subj_type)
            obj_node_id = self._find_or_create_entity_direct(obj, chunk_id, obj_type)

            self.graph.add_edge(subj_node_id, obj_node_id, relation=pred)

    def process_level1_level2_agent(self, chunk: str, id: int):
        """Process attributes (level 1) and triples (level 2) with agent mechanism for schema evolution."""
        prompt = self._get_construction_prompt(chunk)
        llm_response = self.extract_with_llm(prompt)

        # Validate and parse response (reuse helper method)
        parsed_response = self._validate_and_parse_llm_response(prompt, llm_response)
        if not parsed_response:
            return

        # Handle schema evolution
        new_schema_types = parsed_response.get("new_schema_types", {})
        if new_schema_types:
            self._update_schema_with_new_types(new_schema_types)

        extracted_attr = parsed_response.get("attributes", {})
        extracted_triples = parsed_response.get("triples", [])
        entity_types = parsed_response.get("entity_types", {})

        with self.lock:
            self._process_attributes_agent(extracted_attr, id, entity_types)
            self._process_triples_agent(extracted_triples, id, entity_types)

    def _update_schema_with_new_types(self, new_schema_types: Dict[str, List[str]]):
        """Update the schema file with new types discovered by the agent"""
        try:
            schema_paths = {
                "hotpot": "schemas/hotpot.json",
                "2wiki": "schemas/2wiki.json",
                "musique": "schemas/musique.json",
                "novel": "schemas/novels_chs.json",
                "graphrag-bench": "schemas/graphrag-bench.json"
            }

            schema_path = schema_paths.get(self.dataset_name)
            if not schema_path:
                return

            with open(schema_path, 'r', encoding='utf-8') as f:
                current_schema = json.load(f)

            updated = False

            if "nodes" in new_schema_types:
                for new_node in new_schema_types["nodes"]:
                    if new_node not in current_schema.get("Nodes", []):
                        current_schema.setdefault("Nodes", []).append(new_node)
                        updated = True

            if "relations" in new_schema_types:
                for new_relation in new_schema_types["relations"]:
                    if new_relation not in current_schema.get("Relations", []):
                        current_schema.setdefault("Relations", []).append(new_relation)
                        updated = True

            if "attributes" in new_schema_types:
                for new_attribute in new_schema_types["attributes"]:
                    if new_attribute not in current_schema.get("Attributes", []):
                        current_schema.setdefault("Attributes", []).append(new_attribute)
                        updated = True

            # Save updated schema back to file
            if updated:
                with open(schema_path, 'w', encoding='utf-8') as f:
                    json.dump(current_schema, f, ensure_ascii=False, indent=2)

                # Update the in-memory schema
                self.schema = current_schema

        except Exception as e:
            logger.error(f"Failed to update schema for dataset '{self.dataset_name}': {type(e).__name__}: {e}")

    def _connect_keywords_to_communities(self):
        """Connect relevant keywords to communities"""
        # comm_names = [self.graph.nodes[n]['properties']['name'] for n, d in self.graph.nodes(data=True) if d['level'] == 4]
        comm_nodes = [n for n, d in self.graph.nodes(data=True) if d['level'] == 4]
        kw_nodes = [n for n, d in self.graph.nodes(data=True) if d['label'] == 'keyword']
        with self.lock:
            for comm in comm_nodes:
                comm_name = self.graph.nodes[comm]['properties']['name'].lower()
                for kw in kw_nodes:
                    kw_name = self.graph.nodes[kw]['properties']['name'].lower()
                    if kw_name in comm_name or comm_name in kw_name:
                        self.graph.add_edge(kw, comm, relation="describes")

    def process_document(self, doc: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Process a single document and return its results."""
        try:
            if not doc:
                raise ValueError("Document is empty or None")

            chunks, chunk2id = self.chunk_text(doc)

            if not chunks or not chunk2id:
                raise ValueError(
                    f"No valid chunks generated from document. Chunks: {len(chunks)}, Chunk2ID: {len(chunk2id)}")

            for chunk in chunks:
                try:
                    id = next(key for key, value in chunk2id.items() if value == chunk)
                except StopIteration:
                    id = nanoid.generate(size=8)
                    chunk2id[id] = chunk

                if self.mode == "agent":
                    self.process_level1_level2_agent(chunk, id)
                else:
                    self.process_level1_level2(chunk, id)

        except Exception as e:
            error_msg = f"Error processing document: {type(e).__name__}: {str(e)}"
            raise Exception(error_msg) from e

    def process_all_documents(self, documents: List[Dict[str, Any]]) -> None:
        """Process all documents with high concurrency and pass results to process_level4."""

        max_workers = min(self.config.construction.max_workers, (os.cpu_count() or 1) + 4)
        start_construct = time.time()
        total_docs = len(documents)

        logger.info(f"Starting processing {total_docs} documents with {max_workers} workers...")

        all_futures = []
        processed_count = 0
        failed_count = 0

        try:
            with futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
                # Submit all documents for processing and store futures
                all_futures = [executor.submit(self.process_document, doc) for doc in documents]

                for i, future in enumerate(futures.as_completed(all_futures)):
                    try:
                        future.result()
                        processed_count += 1

                        if processed_count % 10 == 0 or processed_count == total_docs:
                            elapsed_time = time.time() - start_construct
                            avg_time_per_doc = elapsed_time / processed_count if processed_count > 0 else 0
                            remaining_docs = total_docs - processed_count
                            estimated_remaining_time = remaining_docs * avg_time_per_doc

                            logger.info(f"Progress: {processed_count}/{total_docs} documents processed "
                                        f"({processed_count / total_docs * 100:.1f}%) "
                                        f"[{failed_count} failed] "
                                        f"ETA: {estimated_remaining_time / 60:.1f} minutes")

                    except Exception as e:
                        failed_count += 1

        except Exception as e:
            return

        end_construct = time.time()
        logger.info(f"Construction Time: {end_construct - start_construct}s")
        logger.info(f"Successfully processed: {processed_count}/{total_docs} documents")
        logger.info(f"Failed: {failed_count} documents")

        logger.info(f"---- {'Processing Level 3 and 4':^20} ----")
        logger.info(f"{'-' * 20}")
        self.triple_deduplicate()
        # self.process_level4() # Process level 4 build community relations

    def triple_deduplicate(self):
        """deduplicate triples in lv1 and lv2"""
        new_graph = nx.MultiDiGraph()

        for node, node_data in self.graph.nodes(data=True):
            new_graph.add_node(node, **node_data)

        seen_keys = set()
        for u, v, key, data in self.graph.edges(keys=True, data=True):
            if (u, v, key) not in seen_keys:
                seen_keys.add((u, v, key))
                new_graph.add_edge(u, v, key=key, **data)
        self.graph = new_graph

    def format_output(self, file_name: str) -> List[Dict[str, Any]]:
        """convert graph to specified output format"""
        output = []

        for u, v, data in self.graph.edges(data=True):
            u_data = self.graph.nodes[u]
            v_data = self.graph.nodes[v]
            u_data["properties"]["file_names"] = [file_name]
            v_data["properties"]["file_names"] = [file_name]

            relationship = {
                "start_node": {
                    "label": u_data["label"],
                    "properties": u_data["properties"],
                },
                "relation": data["relation"],
                "end_node": {
                    "label": v_data["label"],
                    "properties": v_data["properties"],
                },
            }
            output.append(relationship)

        return output

    def save_graphml(self, output_path: str):
        graph_processor.save_graph(self.graph, output_path)

    def build_knowledge_graph(self, file_name: str, documents: List[Dict]):
        logger.info(f"========{'Start Building':^20}========")
        logger.info(f"{'-' * 30}")

        self.process_all_documents(documents)

        logger.info(f"All Process finished, token cost: {self.token_len}")

        # self.save_chunks_to_file() # Save chunked results to file

        output = self.format_output(file_name)

        # json_output_path = f"output/graphs/{self.dataset_name}_new.json"
        # os.makedirs("output/graphs", exist_ok=True)
        # with open(json_output_path, 'w', encoding='utf-8') as f:
        #     json.dump(output, f, ensure_ascii=False, indent=2)
        # logger.info(f"Graph saved to {json_output_path}")

        return output

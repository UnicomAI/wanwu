import json
import os
from pathlib import Path
import networkx as nx

from utils.logger import logger
from utils.community_reports import CommunityReportsExtractor


GRAPH_FIELD_SEP = "<SEP>"

def generate_subgraph(relationships: list) -> nx.MultiDiGraph:
    """
    generate a sub knowledge graph

    Expected JSON format:
    [
        {
            "start_node": {
                "label": "entity",
                "properties": {"name": "Entity Name", "description": "..."}
            },
            "relation": "relation_type",
            "end_node": {
                "label": "entity",
                "properties": {"name": "Entity Name", "description": "..."}
            }
        }
    ]
    """
    graph = nx.MultiDiGraph()
    node_mapping = {}  # (name, schema_type) -> node_id

    for rel in relationships:
        start_node_data = rel["start_node"]
        end_node_data = rel["end_node"]
        relation = rel["relation"]

        # Create unique key for start node - ensure name is a string
        start_name = start_node_data["properties"].get("name", "")
        if isinstance(start_name, list):
            start_name = ", ".join(str(item) for item in start_name)
        elif not isinstance(start_name, str):
            start_name = str(start_name)

        schema_type = start_node_data["properties"].get("schema_type", "")
        start_key = (start_name, schema_type)
        if start_key not in node_mapping:
            node_id = start_name
            # if schema_type:
            #     node_id = f"{start_name}_{schema_type}"
            node_mapping[start_key] = node_id

            # Add node with all properties
            node_attrs = {
                "label": start_node_data["label"],
                "properties": start_node_data["properties"],
                "description": "",
            }

            # Add level based on label
            if start_node_data["label"] == "attribute":
                node_attrs["level"] = 1
            elif start_node_data["label"] == "entity":
                node_attrs["level"] = 2
            elif start_node_data["label"] == "keyword":
                node_attrs["level"] = 3
            elif start_node_data["label"] == "community":
                node_attrs["level"] = 4
            else:
                node_attrs["level"] = 2  # Default level

            graph.add_node(node_id, **node_attrs)

        # Create unique key for end node - ensure name is a string
        end_name = end_node_data["properties"].get("name", "")
        if isinstance(end_name, list):
            end_name = ", ".join(str(item) for item in end_name)
        elif not isinstance(end_name, str):
            end_name = str(end_name)

        schema_type = end_node_data["properties"].get("schema_type", "")
        end_key = (end_name, schema_type)
        if end_key not in node_mapping:
            node_id = end_name
            # if schema_type:
            #     node_id = f"{end_name}_{schema_type}"
            node_mapping[end_key] = node_id

            # Add node with all properties
            node_attrs = {
                "label": end_node_data["label"],
                "properties": end_node_data["properties"],
                "description": "",
            }

            # Add level based on label
            if end_node_data["label"] == "attribute":
                node_attrs["level"] = 1
            elif end_node_data["label"] == "entity":
                node_attrs["level"] = 2
            elif end_node_data["label"] == "keyword":
                node_attrs["level"] = 3
            elif end_node_data["label"] == "community":
                node_attrs["level"] = 4
            else:
                node_attrs["level"] = 2  # Default level

            graph.add_node(node_id, **node_attrs)

        # Add edge
        start_id = node_mapping[start_key]
        end_id = node_mapping[end_key]
        graph.add_edge(start_id, end_id, relation=relation)

    return graph


def load_graph_from_json(input_path: str) -> nx.MultiDiGraph:
    """
    Load a knowledge graph from JSON format
    
    Expected JSON format:
    [
        {
            "start_node": {
                "label": "entity",
                "properties": {"name": "Entity Name", "description": "..."}
            },
            "relation": "relation_type",
            "end_node": {
                "label": "entity", 
                "properties": {"name": "Entity Name", "description": "..."}
            }
        }
    ]
    """
    with open(input_path, 'r', encoding='utf-8') as f:
        relationships = json.load(f)

    return generate_subgraph(relationships)


def save_graph_to_json(graph: nx.MultiDiGraph, output_path: str):
    """
    Save a knowledge graph to JSON format
    
    Output format:
    [
        {
            "start_node": {
                "label": "entity",
                "properties": {"name": "Entity Name", "description": "..."}
            },
            "relation": "relation_type", 
            "end_node": {
                "label": "entity",
                "properties": {"name": "Entity Name", "description": "..."}
            }
        }
    ]
    """
    output = []
    
    for u, v, data in graph.edges(data=True):
        u_data = graph.nodes[u]
        v_data = graph.nodes[v]
        
        relationship = {
            "start_node": {
                "label": u_data["label"],
                "properties": u_data["properties"],
                "description": u_data["description"],
            },
            "relation": data["relation"],
            "end_node": {
                "label": v_data["label"],
                "properties": v_data["properties"],
                "description": u_data["description"],
            },
        }
        output.append(relationship)

    # Make sure the directory for the output path exists
    output_dir = Path(output_path).parent
    output_dir.mkdir(parents=True, exist_ok=True)

    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(output, f, ensure_ascii=False, indent=2)


# Legacy function for backward compatibility
def load_graph(input_path: str) -> nx.MultiDiGraph:
    """
    Load graph from either JSON or GraphML format (legacy support)
    """
    if input_path.endswith('.json'):
        return load_graph_from_json(input_path)
    elif input_path.endswith('.graphml'):
        return load_graph_from_graphml(input_path)
    else:
        raise ValueError(f"Unsupported file format: {input_path}")


def load_graph_from_graphml(input_path: str) -> nx.MultiDiGraph:
    """
    Load graph from GraphML format (legacy function)
    """
    graph_data = nx.read_graphml(input_path)
    
    for node_id, data in graph_data.nodes(data=True):
        # Handle properties (d1)
        if "d1" in data:
            try:
                data["properties"] = json.loads(data["d1"])
                del data["d1"]
            except json.JSONDecodeError:
                logger.warning(f"Warning: Could not parse properties for node {node_id}")
                data["properties"] = {"name": str(data["d1"])}
                del data["d1"]
        
        # Handle level (d2)
        if "d2" in data:
            try:
                data["level"] = int(data["d2"])
                del data["d2"]
            except (ValueError, TypeError):
                data["level"] = 2  # Default level if conversion fails
                del data["d2"]
        
        # Handle label (d0)
        if "d0" in data:
            data["label"] = str(data["d0"])
            del data["d0"]
    
    for u, v, data in graph_data.edges(data=True):
        # Handle relation (d3)
        if "d3" in data:
            data["relation"] = str(data["d3"]).strip('"')
            del data["d3"]
    
    return graph_data


def save_graph(graph: nx.MultiDiGraph, output_path: str):
    """
    Save graph to either JSON or GraphML format based on file extension
    """
    if output_path.endswith('.json'):
        save_graph_to_json(graph, output_path)
    elif output_path.endswith('.graphml'):
        save_graph_to_graphml(graph, output_path)
    else:
        raise ValueError(f"Unsupported output format: {output_path}")


def save_graph_to_graphml(graph: nx.MultiDiGraph, output_path: str):
    """
    Save graph to GraphML format (legacy function)
    """
    # Create a copy of the graph to avoid modifying the original
    graph_copy = graph.copy()
    
    for n, data in graph_copy.nodes(data=True):
        for k, v in list(data.items()):  
            if isinstance(v, dict):
                graph_copy.nodes[n][k] = json.dumps(v, ensure_ascii=False)

    for u, v, data in graph_copy.edges(data=True):
        for k, v in list(data.items()):
            if isinstance(v, dict):
                graph_copy.edges[u, v][k] = json.dumps(v, ensure_ascii=False)

    nx.write_graphml(graph_copy, output_path)


def delete_file(user_id: str, kb_name: str, file_name: str):
    to_remove_nodes = []
    graph = None

    file_path = Path(file_name)
    graph_path = f"./data/graph/{user_id}/{kb_name}.json"
    if os.path.exists(graph_path):
        graph = load_graph(graph_path)

    if graph is not None:
        for node_name, attr in graph.nodes(data=True):
            file_names = attr['properties']['file_names']
            file_names = [x for x in file_names if x != file_name]
            attr['properties']['file_names'] = file_names
            if not file_names:
                to_remove_nodes.append(node_name)

        for node_name in to_remove_nodes:
            graph.remove_node(node_name)

        for node_degree in graph.degree:
            graph.nodes[str(node_degree[0])]["rank"] = int(node_degree[1])

        save_graph(graph, graph_path)

    return graph

def delete_kb(user_id: str, kb_name: str):
    graph_path = f"./data/graph/{user_id}/{kb_name}.json"
    if os.path.exists(graph_path):
        os.remove(graph_path)


def graph_merge(g1: nx.MultiDiGraph, g2: nx.MultiDiGraph):
    """Merge graph g2 into g1 in place."""
    for node_name, attr in g2.nodes(data=True):
        if not g1.has_node(node_name):
            g1.add_node(node_name, **attr)
            continue
        file_names = g1.nodes[node_name]['properties']['file_names']
        g1.nodes[node_name]['properties']['file_names'] = list(set(file_names + attr['properties']['file_names']))

    for source, target, attr in g2.edges(data=True):
        edge = g1.get_edge_data(source, target)
        if edge is None:
            g1.add_edge(source, target, **attr)
            continue

    for node_degree in g1.degree:
        g1.nodes[str(node_degree[0])]["rank"] = int(node_degree[1])

    return g1


def merge_subgraph(
    subgraph: nx.MultiDiGraph,
    old_graph_path: str
):
    # Check if the file exists
    if os.path.exists(old_graph_path):
        old_graph = load_graph(old_graph_path)
        if old_graph is not None:
            logger.info("Merge with an exiting graph...................")
            new_graph = graph_merge(old_graph, subgraph)
        else:
            new_graph = subgraph
    else:
        new_graph = subgraph
    pr = nx.pagerank(new_graph)
    for node_name, pagerank in pr.items():
        new_graph.nodes[node_name]["pagerank"] = pagerank

    save_graph(new_graph, old_graph_path)

    return new_graph


def extract_community(graph, config):
    ext = CommunityReportsExtractor(config)
    cr = ext(graph)
    community_structure = cr.structured_output
    community_reports = cr.output

    reports = []
    for structure, rep in zip(community_structure, community_reports):
        obj = {
            "report": rep,
            "entities": structure["entities"],
            "report_title": structure["title"],
        }
        reports.append(obj)

    return reports


def update_graph(user_id:str, kb_name: str, file_name: str, relationships: list):
    # =========== Generate subgraph =============
    subgraph = generate_subgraph(relationships)

    # =========== Merge subgraph =============
    new_file_path = f"./data/graph/{user_id}/{kb_name}.json"
    new_graph = merge_subgraph(subgraph, new_file_path)

    return new_graph

if __name__ == "__main__":
    relationships = [{'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '文物年份：明清时期及以后制作'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '种类：国王面具、王后面具、活佛面具、神仙面具、动物面具等'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '材质：布料、皮革、木材'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '制作工艺：绘画、雕刻、缝制'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '色彩特征：鲜艳，对比强烈'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '颜色象征：白色代表纯洁善良，红色代表威严勇猛，黑色代表邪恶'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '造型特点：五官夸张，突出角色特征'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '功能用途：用于藏戏表演'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '历史时期：历史悠久，现存多为明清及以后制作'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '包含类型',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '国王面具', 'schema_type': '面具类型'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '包含类型',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '王后面具', 'schema_type': '面具类型'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '包含类型',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '活佛面具', 'schema_type': '面具类型'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '包含类型',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '神仙面具', 'schema_type': '面具类型'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '包含类型',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '动物面具', 'schema_type': '面具类型'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '制作材质',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '布料', 'schema_type': '材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '制作材质',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '皮革', 'schema_type': '材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '制作材质',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '木材', 'schema_type': '材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '制作工艺',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '绘画', 'schema_type': '工艺'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '制作工艺',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '雕刻', 'schema_type': '工艺'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '制作工艺',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '缝制', 'schema_type': '工艺'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '颜色象征',
                      'end_node': {'label': 'entity', 'properties': {'name': '白色代表纯洁善良'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '颜色象征',
                      'end_node': {'label': 'entity', 'properties': {'name': '红色代表威严勇猛'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '颜色象征',
                      'end_node': {'label': 'entity', 'properties': {'name': '黑色代表邪恶'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '造型特点',
                      'end_node': {'label': 'entity', 'properties': {'name': '五官夸张'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '功能用途',
                      'end_node': {'label': 'entity', 'properties': {'name': '用于藏戏表演'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '主要制作时期',
                      'end_node': {'label': 'entity', 'properties': {'name': '明清时期'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '出土地点：西藏自治区、青海、甘肃、四川等地的藏区'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '出土时间：不同时期陆续发现'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '材质：布料、皮革、木材'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '工艺：绘画、雕刻、缝制'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '色彩象征：白色代表纯洁善良，红色代表威严勇猛，黑色代表邪恶'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '造型特征：五官夸张，突出角色特点'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '角色类型：国王、王后、活佛、神仙、动物等'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '明代织锦和刺绣工艺', 'schema_type': '工艺技术'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '工艺水平：高超'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '明代织锦和刺绣工艺', 'schema_type': '工艺技术'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '所属时期：明代'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '明代织锦和刺绣工艺', 'schema_type': '工艺技术'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '应用领域：藏传佛教宗教仪式、审美表达'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏传佛教密宗文化', 'schema_type': '宗教文化'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '文化属性：宗教与艺术融合'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏传佛教密宗文化', 'schema_type': '宗教文化'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '研究价值：重要文物佐证'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏艺术', 'schema_type': '艺术形式'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '组成部分：藏戏面具'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏艺术', 'schema_type': '艺术形式'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '文化地位：藏民族文化的瑰宝'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏艺术', 'schema_type': '艺术形式'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '功能：体现宗教信仰、民俗风情、审美观念、艺术创造力'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '体现',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '藏戏艺术', 'schema_type': '艺术形式'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '反映',
                      'end_node': {'label': 'entity', 'properties': {'name': '藏民族宗教信仰'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '反映',
                      'end_node': {'label': 'entity', 'properties': {'name': '民俗风情'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '反映',
                      'end_node': {'label': 'entity', 'properties': {'name': '审美观念'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '反映',
                      'end_node': {'label': 'entity', 'properties': {'name': '艺术创造力'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '用于',
                      'end_node': {'label': 'entity', 'properties': {'name': '藏戏表演'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '制作于',
                      'end_node': {'label': 'entity', 'properties': {'name': '明清时期'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '出土于',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '西藏自治区', 'schema_type': '地理区域'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '出土于',
                      'end_node': {'label': 'entity', 'properties': {'name': '青海藏区'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '出土于',
                      'end_node': {'label': 'entity', 'properties': {'name': '甘肃藏区'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '出土于',
                      'end_node': {'label': 'entity', 'properties': {'name': '四川藏区'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '材质包括',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '布料', 'schema_type': '材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '材质包括',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '皮革', 'schema_type': '材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '材质包括',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '木材', 'schema_type': '材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '通过工艺',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '绘画', 'schema_type': '工艺'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '通过工艺',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '雕刻', 'schema_type': '工艺'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '通过工艺',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '缝制', 'schema_type': '工艺'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '象征意义',
                      'end_node': {'label': 'entity', 'properties': {'name': '白色代表纯洁善良'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '象征意义',
                      'end_node': {'label': 'entity', 'properties': {'name': '红色代表威严勇猛'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '藏戏面具', 'schema_type': '文物'}},
                      'relation': '象征意义',
                      'end_node': {'label': 'entity', 'properties': {'name': '黑色代表邪恶'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '明代织锦和刺绣工艺', 'schema_type': '工艺技术'}},
                      'relation': '体现于',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '反映',
                      'end_node': {'label': 'entity', 'properties': {'name': '藏传佛教宗教仪式'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '反映',
                      'end_node': {'label': 'entity', 'properties': {'name': '信仰内涵'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '反映',
                      'end_node': {'label': 'entity', 'properties': {'name': '审美观念'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '用于研究',
                      'end_node': {'label': 'entity', 'properties': {'name': '明代藏传佛教文化'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '用于研究',
                      'end_node': {'label': 'entity', 'properties': {'name': '宗教艺术'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '用于研究',
                      'end_node': {'label': 'entity', 'properties': {'name': '服饰文化'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '用于研究',
                      'end_node': {'label': 'entity', 'properties': {'name': '工艺技术'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '属于',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '藏传佛教密宗文化', 'schema_type': '宗教文化'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '西藏珍贵文物', 'schema_type': '文物'}},
                      'relation': '是',
                      'end_node': {'label': 'entity', 'properties': {'name': '重要文物'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '材质：铁质'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '形状：正方形'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '边长：约12.5厘米'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '印钮：龙钮'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '纹饰特征：龙身蜿蜒，形态矫健'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '文字：八思巴文'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '铭文内容：灌顶国师阐化王之印'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '字体特征：线条刚劲有力，结构规整'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '保存状况：整体保存较为完好'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '文物年份：明代'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '出土时间：无明确记载'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '出土地点：无明确出土地点'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '收藏状态：现藏于相关博物馆或收藏机构'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '画面内容：大慈法王形象'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {
                                       'name': '人物特征：面容慈祥，身着华丽僧袍，头戴僧帽，双手结印，端坐莲花宝座'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '背景元素：祥云、花草树木、佛教法器'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '艺术风格：色彩丰富，针法细腻'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '工艺特征：刺绣精湛，边缘以锦缎装裱'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '视觉效果：色彩鲜艳，历经数百年仍光彩夺目'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '文物年份：明代'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '出土时间：无明确出土时间'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '出土地点：无明确出土地点'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '流传范围：西藏及其他藏传佛教传播地区'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '造型：扇形'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '结构：由五片莲瓣组成'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '每片装饰：绣有一尊佛像'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '佛像特征：神态庄严，法相慈悲'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '主要工艺：织锦夹金'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '材质特征：丝线细腻，含金线'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '视觉效果：光线照耀下熠熠生辉'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '附加装饰：宝石、珍珠'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '边缘工艺：精美刺绣'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '图案内容：祥云、莲花等吉祥图案'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '文物年份：明代'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '出土时间：无明确出土时间'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute', 'properties': {'name': '出土地点：无明确出土地点'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': 'has_attribute',
                      'end_node': {'label': 'attribute',
                                   'properties': {'name': '流传区域：藏传佛教寺庙或相关宗教文化区域'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '属于朝代',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '明代', 'schema_type': '朝代'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '材质为',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '铁质', 'schema_type': '材质'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '印钮为',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '龙钮', 'schema_type': '器物部件'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '铭文使用文字',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '八思巴文', 'schema_type': '文字'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '铭文内容是',
                      'end_node': {'label': 'entity', 'properties': {'name': '灌顶国师阐化王之印'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '反映历史制度',
                      'end_node': {'label': 'entity', 'properties': {'name': '明朝对西藏的册封制度'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '见证',
                      'end_node': {'label': 'entity', 'properties': {'name': '明朝与西藏政治关系'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '具有历史价值',
                      'end_node': {'label': 'entity', 'properties': {'name': '研究明朝民族政策'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '灌顶国师阐化王印', 'schema_type': '文物'}},
                      'relation': '具有历史价值',
                      'end_node': {'label': 'entity', 'properties': {'name': '研究边疆治理'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': '属于朝代',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '明代', 'schema_type': '朝代'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': '描绘对象',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '大慈法王', 'schema_type': '宗教人物'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': '艺术形式',
                      'end_node': {'label': 'entity', 'properties': {'name': '刺绣唐卡'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': '体现艺术水平',
                      'end_node': {'label': 'entity', 'properties': {'name': '明代藏传佛教艺术高超水平'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': '反映宗教信仰',
                      'end_node': {'label': 'entity', 'properties': {'name': '对大慈法王的尊崇'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': '具有价值',
                      'end_node': {'label': 'entity', 'properties': {'name': '研究藏传佛教艺术'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '大慈法王刺绣唐卡', 'schema_type': '文物'}},
                      'relation': '具有价值',
                      'end_node': {'label': 'entity', 'properties': {'name': '研究藏族刺绣工艺'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '属于朝代',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '明代', 'schema_type': '朝代'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '结构组成',
                      'end_node': {'label': 'entity', 'properties': {'name': '五片莲瓣'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '每片装饰',
                      'end_node': {'label': 'entity', 'properties': {'name': '一尊佛像'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '工艺技术',
                      'end_node': {'label': 'entity', 'properties': {'name': '织锦夹金'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '装饰材料',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '金线', 'schema_type': '材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '附加装饰',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '宝石', 'schema_type': '装饰材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '附加装饰',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '珍珠', 'schema_type': '装饰材料'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '边缘图案',
                      'end_node': {'label': 'entity',
                                   'properties': {'name': '祥云', 'schema_type': '图案元素'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '边缘图案',
                      'end_node': {'label': 'entity', 'properties': {'name': '莲花'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '用途为',
                      'end_node': {'label': 'entity', 'properties': {'name': '藏传佛教密宗法器'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '反映内容',
                      'end_node': {'label': 'entity', 'properties': {'name': '宗教仪式'}}},
                     {'start_node': {'label': 'entity',
                                     'properties': {'name': '织锦夹金五佛冠', 'schema_type': '文物'}},
                      'relation': '反映内容',
                      'end_node': {'label': 'entity', 'properties': {'name': '审美观念'}}}]
    sub_graph = load_graph_from_json(relationships)
    print(sub_graph)

import os
import sys
import json
import asyncio
import glob
import shutil
import copy
from typing import List, Dict, Optional
import time
from datetime import datetime

# Add project root to path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

# FastAPI imports
from fastapi import FastAPI, UploadFile, File, HTTPException, WebSocket, WebSocketDisconnect, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel
import uvicorn

from utils.logger import logger
import ast
from utils import kt_gen as constructor
from config import get_config, ConfigManager, prompt_templates
from utils import graph_processor


app = FastAPI(title="graph Unified Interface", version="1.0.0")

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Global variables
active_connections: Dict[str, WebSocket] = {}

CONFIG = get_config()



class ConnectionManager:
    def __init__(self):
        self.active_connections: Dict[str, WebSocket] = {}

    async def connect(self, websocket: WebSocket, client_id: str):
        await websocket.accept()
        self.active_connections[client_id] = websocket

    def disconnect(self, client_id: str):
        if client_id in self.active_connections:
            del self.active_connections[client_id]

    async def send_message(self, message: dict, client_id: str):
        if client_id in self.active_connections:
            try:
                await self.active_connections[client_id].send_text(json.dumps(message))
            except Exception as e:
                logger.error(f"Error sending message to {client_id}: {e}")
                self.disconnect(client_id)


manager = ConnectionManager()


# Request/Response models
class ExtracGraphDataResponse(BaseModel):
    """ res_data 格式字段"""
    success: bool
    message: str
    graph_chunks: List[Dict] = []
    graph_vocabulary_set: set = set()
    community_reports: List[Dict] = []


# Request/Response models
class CommunityReportsResponse(BaseModel):
    """ res_data 格式字段"""
    success: bool
    message: str
    community_reports: List[Dict] = []

# Request/Response models
class RequestResponse(BaseModel):
    """ res_data 格式字段"""
    success: bool
    message: str

async def send_progress_update(client_id: str, stage: str, progress: int, message: str):
    """Send progress update via WebSocket"""
    await manager.send_message({
        "type": "progress",
        "stage": stage,
        "progress": progress,
        "message": message,
        "timestamp": datetime.now().isoformat()
    }, client_id)


@app.post("/api/extrac_graph_data", response_model=ExtracGraphDataResponse)
async def extrac_graph_data(request: Request):
    """extrac_graph_data endpoint  chunks: List[Dict], client_id: str = 'default' """
    try:
        json_request = await request.json()
        chunks = json_request["chunks"]
        user_id = json_request["user_id"]
        file_name = json_request["file_name"]
        kb_name = json_request["kb_name"]
        llm_model = json_request["llm_model"]
        llm_base_url = json_request["llm_base_url"]
        llm_api_key = json_request["llm_api_key"]
        temperature = json_request.get("temperature", 0.001)
        for chunk in chunks:
            chunk["old_snippet"] = chunk["snippet"]
            chunk["snippet"] = f"{file_name.split('.')[0]}:" + chunk["snippet"]
        client_id = json_request.get("client_id", "default")
        schema = json_request.get("schema", None)
        config = get_config()
        config.construction.mode = "general"  # "agent"
        dataset = "demo"
        dataset_config = config.get_dataset_config(dataset)
        dataset_config.corpus_path = "data/demo/custom_corpus.json"
        dataset_config.schema_path = "schemas/custom.json"
        dataset_config.graph_output = "output/graphs/custom_new.json"
        config.prompts["construction"]["general"] = prompt_templates.general_zh_prompt_template
        config.construction.LLM_MODEL = llm_model
        config.construction.LLM_BASE_URL = llm_base_url
        config.construction.LLM_API_KEY = llm_api_key
        config.construction.TEMPERATURE = temperature
        embedding_model = None
        res_data = []
        builder = constructor.KTBuilder(
            dataset,
            embedding_model,
            dataset_config.schema_path,
            schema=schema,
            mode=config.construction.mode,
            config=config
        )
        res_data = builder.build_knowledge_graph(file_name, chunks)

        # =========== 更新 graph ============= [EN] =========== Update graph =============
        graph_processor.update_graph(user_id, kb_name, file_name, res_data)

        # =========== 整理 graph_vocabulary_set ============= [EN] =========== Organize graph_vocabulary_set =============
        graph_vocabulary_set = set()
        for node in builder.graph.nodes:
            node_json = builder.graph.nodes[node]
            # print(node_json)
            if node_json['properties'].get('schema_type'):
                schema_type = f"K:{node_json['properties'].get('schema_type')}"
            else:
                schema_type = "K:graph_node"
            node_msg = f"{node_json['properties']['name']}|||schema_type:{schema_type}"
            graph_vocabulary_set.add(node_msg)
        # =========== 整理 graph_vocabulary_set ============= [EN] =========== Organize graph_vocabulary_set =============

        # =========== 整理 graph_chunks start============= [EN] =========== Organize graph_chunks start==============
        graph_chunks = []
        for triple in res_data:
            reference_chunk_id = triple["start_node"]["properties"]["chunk id"]
            meta_data = builder.all_chunks[reference_chunk_id]["meta_data"]
            meta_data["reference_snippet"] = builder.all_chunks[reference_chunk_id]["old_snippet"]
            temp_triple = copy.deepcopy(triple)
            # 移除 start_node 中的 'chunk id' [EN] Remove 'chunk id' in start_node
            if 'chunk id' in temp_triple['start_node']['properties']:
                del temp_triple['start_node']['properties']['chunk id']
            # 移除 end_node 中的 'chunk id' [EN] Remove 'chunk id' in end_node
            if 'chunk id' in temp_triple['end_node']['properties']:
                del temp_triple['end_node']['properties']['chunk id']
            graph_data_text = f"{triple['start_node']['properties']['name']} {triple['relation']} {triple['end_node']['properties']['name']}"
            # print(graph_data_text)
            graph_chunks.append(
                {"chunk_type": "graph", "graph_data_text": graph_data_text, "graph_data": copy.deepcopy(temp_triple),
                 "meta_data": meta_data})
        # =========== 整理 graph_chunks  end ============= [EN] =========== Organize graph_chunks end =============
        # await send_progress_update(client_id, "extrac_graph_data", 10, "extrac_graph_data completed successfully!")

        return ExtracGraphDataResponse(
            success=True,
            message="Files uploaded successfully",
            graph_chunks=graph_chunks,
            graph_vocabulary_set=graph_vocabulary_set,
        )

    except Exception as e:
        # await send_progress_update(client_id, "extrac_graph_data", 0, f"Upload failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/generate_community_reports", response_model=ExtracGraphDataResponse)
async def generate_community_reports(request: Request):
    """extrac_graph_data endpoint  chunks: List[Dict], client_id: str = 'default' """
    try:
        json_request = await request.json()
        # relationships = json_request["graph_data"]
        user_id = json_request["user_id"]
        kb_name = json_request["kb_name"]
        llm_model = json_request["llm_model"]
        llm_base_url = json_request["llm_base_url"]
        llm_api_key = json_request["llm_api_key"]
        # file_name = json_request["file_name"]
        client_id = json_request.get("client_id", "default")
        logger.info(f"generate_community_reports, user_id: {user_id}, kb_name: {kb_name}")
        config = get_config()
        config.construction.mode = "general"  # "agent"
        config.construction.LLM_MODEL = llm_model
        config.construction.LLM_BASE_URL = llm_base_url
        config.construction.LLM_API_KEY = llm_api_key

        # =========== 生成社区报告 ============= [EN] =========== Generate community report =============
        file_path = f"./data/graph/{user_id}/{kb_name}.json"
        new_graph = graph_processor.load_graph_from_json(file_path)
        reports = []
        if os.path.exists(file_path):
            reports = graph_processor.extract_community(new_graph, config)
        await send_progress_update(client_id, "generate_community_reports", 10, "generate_community_reports completed successfully!")

        return CommunityReportsResponse(
            success=True,
            message="Files uploaded successfully",
            community_reports=reports,
        )

    except Exception as e:
        await send_progress_update(client_id, "generate_community_reports", 0, f"Upload failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/delete_file", response_model=ExtracGraphDataResponse)
async def delete_file(request: Request):
    """update graph """
    try:
        json_request = await request.json()
        user_id = json_request["user_id"]
        kb_name = json_request["kb_name"]
        file_name = json_request["file_name"]
        client_id = json_request.get("client_id", "default")

        # =========== 更新 graph ============= [EN] =========== Update graph =============
        graph_processor.delete_file(user_id, kb_name, file_name)
        await send_progress_update(client_id, "delete_file", 10, "delete_file completed successfully!")

        return RequestResponse(
            success=True,
            message="Files deleted successfully",
        )

    except Exception as e:
        await send_progress_update(client_id, "delete_file", 0, f"deleted failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/delete_kb", response_model=ExtracGraphDataResponse)
async def delete_kb(request: Request):
    """update graph """
    try:
        json_request = await request.json()
        user_id = json_request["user_id"]
        kb_name = json_request["kb_name"]
        client_id = json_request.get("client_id", "default")

        # =========== 更新 graph ============= [EN] =========== Update graph =============
        graph_processor.delete_kb(user_id, kb_name)
        await send_progress_update(client_id, "delete_kb", 10, "delete_file completed successfully!")

        return RequestResponse(
            success=True,
            message="delete_kb successfully",
        )

    except Exception as e:
        await send_progress_update(client_id, "delete_kb", 0, f"deleted failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/test_post")
async def test_post(request: Request):
    """test post"""
    json_request = await request.json()
    time.sleep(6)
    return {
        "code": 0,
        "success": True,
        "message": f"test: {json_request} successfully",
    }

class GraphDataResponse(BaseModel):
    """ res_data 格式字段"""
    success: bool
    message: str
    graph_data: dict

@app.post("/api/get_kb_graph_data", response_model=GraphDataResponse)
async def get_kb_graph_data(request: Request):
    """get graph data"""
    try:
        json_request = await request.json()
        user_id = json_request["user_id"]
        kb_name = json_request["kb_name"]
        kb_id = json_request["kb_id"]
        client_id = json_request.get("client_id", "default")

        file_path = f"./data/graph/{user_id}/{kb_name}.json"
        graph_data = {
            "graph": {
                "directed": False,
                "multigraph": False,
                "nodes": [],
                "edges": []
            }
        }

        if os.path.exists(file_path):
            graph = graph_processor.load_graph(file_path)
            for node in graph.nodes(data=True):
                if node[1]["label"] == "entity":
                    graph_data["graph"]["nodes"].append({
                        "entity_name": node[0],
                        "entity_type": node[1]["label"],
                        "description": "",
                        "source_id": node[1]["properties"]["file_names"]
                    })
            for edge in graph.edges(data=True):
                if edge[2]["relation"] != "has_attribute":
                    graph_data["graph"]["edges"].append({
                        "source_entity": edge[0],
                        "target_entity": edge[1],
                        "description": edge[2]["relation"],
                        "weight": 1.0,
                    })

        await send_progress_update(client_id, "delete_kb", 10, "get_kb_graph_data completed successfully!")

        return GraphDataResponse(
            success=True,
            message="get_kb_graph_data successfully",
            graph_data=graph_data
        )

    except Exception as e:
        await send_progress_update(client_id, "get_kb_graph_data", 0, f"deleted failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/test")
async def test():
    """test"""
    return {
        "code": 0,
        "success": True,
        "message": f"test successfully",
    }


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=20050)

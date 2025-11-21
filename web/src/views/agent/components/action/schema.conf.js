export const schemaConfig = {
    json:
        '{\n' +
        '            "openapi": "3.0.0",\n' +
        '            "info":\n' +
        '                {\n' +
        '                    "title": "Seniverse Weather API",\n' +
        '                    "version": "1.0.0",\n' +
        '                    "description": "API providing current weather info, including temperature, conditions, etc."\n' +
        '                },\n' +
        '            "servers":\n' +
        '                [\n' +
        '                    {"url": "https://api.seniverse.com/v3"}\n' +
        '                ],\n' +
        '            "paths":\n' +
        '                {\n' +
        '                    "/weather/now.json": {\n' +
        '                        "get": {\n' +
        '                            "summary": "Weather Query Tool",\n' +
        '                            "operationId": "getWeatherNow",\n' +
        '                            "description": "Get current weather for a location, including temperature and conditions.",\n' +
        '                            "parameters": [{\n' +
        '                                "name": "location",\n' +
        '                                "description": "Query location, can be city name, zip code, etc.",\n' +
        '                                "in": "query",\n' +
        '                                "required": true,\n' +
        '                                "schema": {"type": "string"}\n' +
        '                            }],\n' +
        '                            "responses": {\n' +
        '                                "200": {\n' +
        '                                    "description": "Successfully retrieved weather info",\n' +
        '                                    "content": {\n' +
        '                                        "application/json": {\n' +
        '                                            "schema": {\n' +
        '                                                "type": "object",\n' +
        '                                                "properties": {\n' +
        '                                                    "location": {"type": "string"},\n' +
        '                                                    "text": {"type": "string"},\n' +
        '                                                    "code": {"type": "string"},\n' +
        '                                                    "temperature": {"type": "string"}\n' +
        '                                                }\n' +
        '                                            }\n' +
        '                                        }\n' +
        '                                    }\n' +
        '                                },\n' +
        '                                "default": {\n' +
        '                                    "description": "Error information when request failed",\n' +
        '                                    "content": {\n' +
        '                                        "application/json": {\n' +
        '                                            "schema": {\n' +
        '                                                "type": "object",\n' +
        '                                                "properties": {"error": {"type": "string"}}\n' +
        '                                            }\n' +
        '                                        }\n' +
        '                                    }\n' +
        '                                }\n' +
        '                            }\n' +
        '                        }\n' +
        '                    }\n' +
        '                }\n' +
        '        }',
    yaml:'openapi: 3.0.0\n' +
        'info:\n' +
        '  title: Seniverse Weather API\n' +
        '  version: 1.0.0\n' +
        '  description: API providing current weather info, including temperature, conditions, etc.\n' +
        'servers:\n' +
        '  - url: https://api.seniverse.com/v3\n' +
        'paths:\n' +
        '  /weather/now.json:\n' +
        '    get:\n' +
        '      summary: Weather Query Tool\n' +
        '      operationId: getWeatherNow\n' +
        '      description: Get current weather for a location, including temperature and conditions.\n' +
        '      parameters:\n' +
        '        - name: location\n' +
        '          description: Query location, can be city name, zip code, etc.\n' +
        '          in: query\n' +
        '          required: true\n' +
        '          schema:\n' +
        '            type: string\n' +
        '      responses:\n' +
        '        \'200\':\n' +
        '          description: Successfully retrieved weather info\n' +
        '          content:\n' +
        '            application/json:\n' +
        '              schema:\n' +
        '                type: object\n' +
        '                properties:\n' +
        '                  location:\n' +
        '                    type: string\n' +
        '                  text:\n' +
        '                    type: string\n' +
        '                  code:\n' +
        '                    type: string\n' +
        '                  temperature:\n' +
        '                    type: string\n' +
        '        default:\n' +
        '          description: Error information when request failed\n' +
        '          content:\n' +
        '            application/json:\n' +
        '              schema:\n' +
        '                type: object\n' +
        '                properties:\n' +
        '                  error:\n' +
        '                    type: string'


}

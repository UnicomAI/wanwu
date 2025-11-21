import os
import json
import pandas as pd
import sys

from logging_config import setup_logging

logger_name = 'rag_schema_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))



def parse_excel_to_schema_json(file_path):
    """
    解析 Excel 文件中的 '类目表' 和 '类目属性表'，输出指定 JSON 结构
    """
    schema = {}
    try:
        # Automatic inference engine using pd.read_excel (supports .xls and .xlsx)
        df_category = pd.read_excel(file_path, sheet_name='类目表')
        df_attribute = pd.read_excel(file_path, sheet_name='类目属性表')
        # Clean column names: remove spaces and newlines
        df_category.columns = df_category.columns.str.strip()
        df_attribute.columns = df_attribute.columns.str.strip()

        # === Analysis Category Table ===
        category_list = []
        for _, row in df_category.iterrows():
            item = {
                "类名": str(row["类名"]).strip() if pd.notna(row["类名"]) else "",
                "类描述": str(row["类描述"]).strip() if pd.notna(row["类描述"]) else ""
            }
            category_list.append(item)

        # === Parse category attribute table ===
        attribute_list = []

        for _, row in df_attribute.iterrows():
            class_name = str(row["类名"]).strip() if pd.notna(row["类名"]) else ""
            attr_name = str(row["属性/关系名"]).strip() if pd.notna(row["属性/关系名"]) else ""

            # Fix description field
            key = (class_name, attr_name)

            desc = str(row["属性/关系说明"]).strip() if pd.notna(row["属性/关系说明"]) else ""

            # Processing alias fields (supports multiple aliases separated by |)
            alias = row["别名(多别名以|隔开)"]
            if pd.isna(alias) or str(alias).strip() == "" or str(alias).lower() == "nan":
                alias_str = ""
            else:
                alias_str = str(alias).strip()

            value_type = str(row["值类型"]).strip() if pd.notna(row["值类型"]) else ""

            attribute_list.append({
                "类名": class_name,
                "属性/关系名": attr_name,
                "属性/关系说明": desc,
                "属性别名(多别名以|隔开)": alias_str,
                "值类型": value_type
            })

        # Build the final JSON structure
        schema = {
            "schema定义": {
                "类目表": category_list,
                "类目属性表": attribute_list
            }
        }
        logger.info("schema:%s" % json.dumps(schema, ensure_ascii=False))
    except Exception as e:
        import traceback
        logger.error(traceback.format_exc())
        logger.error(f"无法读取Excel文件或工作表不存在: {e}")
    return schema


if __name__ == "__main__":

    excel_file = "graph_schema_文物.xlsx"  # or "Graph schema definition template.xls"

    try:
        schema_json = parse_excel_to_schema_json(excel_file)
        # Output formatted JSON to console
        print(json.dumps(schema_json, ensure_ascii=False, indent=4))

        # (Optional) Save to file
        # with open("schema_output.json", "w", encoding="utf-8") as f:
        #     json.dump(schema_json, f, ensure_ascii=False, indent=4)
        # print("✅ saved to schema_output.json")

    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)

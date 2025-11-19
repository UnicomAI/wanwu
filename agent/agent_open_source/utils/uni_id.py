
import uuid

def generate_unique_id():
    unique_id = str(uuid.uuid4())  # 生成一个带破折号的UUID [EN] Generate a UUID with a dash
    return unique_id.replace('-', '')  # 移除破折号 [EN] remove dash

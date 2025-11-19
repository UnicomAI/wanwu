
import uuid

def generate_unique_id():
    unique_id = str(uuid.uuid4())  # Generate a UUID with a dash
    return unique_id.replace('-', '')  # remove dash

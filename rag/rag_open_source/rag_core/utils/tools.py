import hashlib

def generate_md5(content_str):
    # Create an md5 hash object
    md5_obj = hashlib.md5()

    # Encode the string because md5 requires bytes type data
    md5_obj.update(content_str.encode('utf-8'))

    # Get the hexadecimal MD5 value
    md5_value = md5_obj.hexdigest()

    return md5_value


import hashlib

def generate_md5(content_str):
    # 创建一个md5 hash对象 [EN] Create an md5 hash object
    md5_obj = hashlib.md5()

    # 对字符串进行编码，因为md5需要bytes类型的数据 [EN] Encode the string because md5 requires bytes type data
    md5_obj.update(content_str.encode('utf-8'))

    # 获取十六进制的MD5值 [EN] Get the hexadecimal MD5 value
    md5_value = md5_obj.hexdigest()

    return md5_value


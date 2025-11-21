import configparser


class Config:
    def __init__(self, filename):
        self.config = configparser.ConfigParser()
        self.config.read(filename)

    def get(self, section, option):
        return self.config.get(section, option)

    def getstr(self, section, option):
        value = self.config.get(section, option)
        if value.startswith('"') and value.endswith('"'):
            return value[1:-1]
        elif value.startswith("'") and value.endswith("'"):
            return value[1:-1]
        else:
            return value

    def getint(self, section, option):
        return self.config.getint(section, option)

    def getfloat(self, section, option):
        return self.config.getfloat(section, option)

    def getboolean(self, section, option):
        """获取布尔值配置项"""
        # Convert string in config file to boolean
        value = self.config.get(section, option)
        if value.lower() in ('yes', 'true', '1'):
            return True
        elif value.lower() in ('no', 'false', '0'):
            return False
        else:
            raise ValueError(f"Option '{option}' in section '{section}' is not a valid boolean value.")

    def getlist(self, section, option):
        """获取布尔值配置项"""
        # Convert string in config file to boolean
        value = self.config.get(section, option)
        return eval(value)


# Usage example
if __name__ == "__main__":
    config = Config('config.ini')

    # Get string configuration
    server = config.get('DEFAULT', 'Server')
    print(f"Server: {server}")

    # Get database configuration
    db_host = config.get('Database', 'Host')
    db_port = config.getint('Database', 'Port')
    print(f"Database Host: {db_host}, Port: {db_port}")

    # Get API configuration
    api_key = config.get('API', 'Key')
    api_endpoint = config.get('API', 'Endpoint')
    print(f"API Key: {api_key}, Endpoint: {api_endpoint}")

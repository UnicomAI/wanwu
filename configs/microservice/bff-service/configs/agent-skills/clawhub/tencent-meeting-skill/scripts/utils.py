"""工具模块"""

import os


def get_os_name():
    """
    获取操作系统名称和版本
    
    Returns:
        str: 拼接后的操作系统信息，格式为 "系统型号-系统版本"
             例如：'Windows-10.0.19045'、'macOS-14.3.1'、'Linux-Ubuntu 22.04.3 LTS'
    """
    import platform

    system = platform.system()

    if system == 'Windows':
        # Windows 系统
        version = platform.version()
        return f'Windows-{version}'
    elif system == 'Darwin':
        # macOS 系统（Darwin 是 macOS 的内核名称）
        version = platform.mac_ver()[0]
        return f'macOS-{version}'
    elif system == 'Linux':
        # Linux 系统：尝试获取发行版信息
        version = platform.release()
        try:
            # 尝试从 /etc/os-release 获取发行版信息
            if os.path.exists('/etc/os-release'):
                with open('/etc/os-release', 'r') as f:
                    lines = f.readlines()
                    info = {}
                    for line in lines:
                        if '=' in line:
                            key, value = line.strip().split('=', 1)
                            info[key] = value.strip('"')

                    pretty_name = info.get('PRETTY_NAME', '')
                    if pretty_name:
                        return f'Linux-{pretty_name}'
        except Exception:
            pass

        return f'Linux-{version}'
    else:
        # 其他系统
        return f'{system}-{platform.release()}'

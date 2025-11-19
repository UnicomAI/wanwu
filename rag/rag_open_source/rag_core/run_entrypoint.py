import argparse

import gunicorn_server
from run import app  # 替换为你的Flask应用导入路径 [EN] Replace with your Flask application import path


def main():
    parser = argparse.ArgumentParser(description='run Flask应用入口点')
    parser.add_argument('--port', type=int, default=5000, help='服务端口号')
    parser.add_argument('--host', type=str, default='0.0.0.0', help='服务主机')
    parser.add_argument('--workers', type=int, default=1, help='启动worker数量')
    parser.add_argument('--timeout', type=int, default=120, help='超时时间（ms）')
    parser.add_argument('--debug', action='store_true', help='调试模式')
    args = parser.parse_args()

    # 记录日志 [EN] logging
    log_file = f"run_logs_{args.port}.log"
    print(f"正在启动Flask应用，端口号为{args.port}...")
    print(f"日志将输出到 {log_file}")

    """运行 Gunicorn 服务器"""
    default_options = {
        'bind': f'{args.host}:{args.port}',
        'workers': args.workers,
        'worker_class': 'sync',
        'timeout': args.timeout,
        'loglevel': 'info',
        'accesslog': log_file,
        'errorlog': log_file,
    }

    gunicorn_app = gunicorn_server.GunicornServer(app.wsgi_app, default_options)
    gunicorn_app.run()


if __name__ == "__main__":
    main()

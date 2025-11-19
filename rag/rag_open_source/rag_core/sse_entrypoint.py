import argparse

import gunicorn_server
from know_sse import app


def main():
    parser = argparse.ArgumentParser(description='sse应用入口点')
    parser.add_argument('--port', type=int, default=10891, help='服务端口号')
    parser.add_argument('--host', type=str, default='0.0.0.0', help='服务主机')
    parser.add_argument('--workers', type=int, default=5, help='工作进程数')
    parser.add_argument('--timeout', type=int, default=600, help='超时时间(秒)')
    args = parser.parse_args()

    # logging
    log_file = f"sse_logs_{args.port}.log"
    print(f"正在启动FastAPI应用，端口号为{args.port}...")
    print(f"日志将输出到 {log_file}")

    """运行 Gunicorn 服务器"""
    default_options = {
        'bind': f'{args.host}:{args.port}',
        'workers': args.workers,
        'worker_class': 'uvicorn.workers.UvicornWorker',
        'timeout': args.timeout,
        'loglevel': 'info',
        'accesslog': log_file,
        'errorlog': log_file,
    }

    gunicorn_app = gunicorn_server.GunicornBaseServer(app, default_options)
    gunicorn_app.run()


if __name__ == "__main__":
    main()

"""TCP-only bridge from ARM64 application containers to the real x64 sandbox."""
import asyncio
import os


async def handle(reader, writer, target_port):
    upstream = None
    try:
        other_reader, upstream = await asyncio.open_connection(
            os.environ.get("SANDBOX_HOST", "finsafe-x64-host"), target_port
        )

        async def copy(source, destination):
            while data := await source.read(65536):
                destination.write(data)
                await destination.drain()
            if destination.can_write_eof():
                destination.write_eof()

        await asyncio.gather(copy(reader, upstream), copy(other_reader, writer))
    except (OSError, ConnectionError):
        pass
    finally:
        writer.close()
        if upstream:
            upstream.close()


async def main():
    servers = []
    for port, target in ((8080, 18080), (4096, 14096), (4097, 14097)):
        async def accept(reader, writer, target_port=target):
            await handle(reader, writer, target_port)

        servers.append(await asyncio.start_server(accept, "0.0.0.0", port))
    print("Sandbox TCP bridge ready; no API or model responses are simulated", flush=True)
    await asyncio.gather(*(server.serve_forever() for server in servers))


asyncio.run(main())

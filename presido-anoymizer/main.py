import asyncio
from anonymizer import AnonymizerService
import grpc
import anonymizer_pb2_grpc
from presidio_analyzer import AnalyzerEngine
from presidio_anonymizer import AnonymizerEngine

from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import BaseModel, IPvAnyInterface


class TLS(BaseModel):
    insecure: bool = True


class Config(BaseSettings):
    model_config = SettingsConfigDict(env_nested_delimiter="__")

    # gRPC server settings
    host: str = "0.0.0.0"
    port: int = 5051
    tls: TLS = TLS()


# 6. Standard boilerplate to boot up the gRPC server
async def serve():

    config = Config()
    # Create an async server
    server = grpc.aio.server()

    # Register your custom class with the gRPC server using the generated helper function
    service = AnonymizerService(
        analyzer_engine=AnalyzerEngine(),  # Use default Presidio AnalyzerEngine
        anonymizer_engine=AnonymizerEngine(),  # Use default Presidio AnonymizerEngine
    )

    anonymizer_pb2_grpc.add_AnonymizerServiceServicer_to_server(service, server)

    # Bind to all interfaces on port 50051
    address = f"{config.host}:{config.port}"
    if config.tls.insecure:
        server.add_insecure_port(address)
    else:
        raise NotImplementedError(
            "TLS is not implemented in this example. Set tls.insecure to true."
        )

    print(f"gRPC Server listening on port {address}...")
    await server.start()

    # Keep the server running
    await server.wait_for_termination()


if __name__ == "__main__":
    # Run the async event loop
    asyncio.run(serve())

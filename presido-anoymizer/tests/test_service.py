import pytest
from unittest.mock import MagicMock
import grpc

# Import your generated proto files
import anonymizer_pb2

# Import the service class you wrote earlier
from anonymizer import AnonymizerService


@pytest.mark.asyncio
async def test_anonymize_batch_success():
    # 1. Initialize the service
    # Note: This will load the Presidio NLP models into memory just like the real server
    service = AnonymizerService()

    # 2. Create a mock gRPC context
    # This simulates the background data gRPC uses (for things like setting error codes)
    mock_context = MagicMock()

    # 3. Create the input request using your generated Protobuf classes
    request = anonymizer_pb2.BatchRequest(
        texts=[
            "My name is John Doe and my phone number is +1-555-1234.",
            "This is a completely safe string with no PII.",
        ],
        language="en",
    )

    # 4. Call your business logic directly
    response = await service.AnonymizeBatch(request, mock_context)

    # 5. Assertions to verify the outcome
    assert len(response.anonymized_texts) == 2

    # Verify PII was redacted (Presidio replaces entities with placeholders like <PERSON>)
    assert "John Doe" not in response.anonymized_texts[0]
    assert "+1-555-1234" not in response.anonymized_texts[0]
    assert "<PERSON>" in response.anonymized_texts[0]

    # Verify safe strings are not altered
    assert (
        response.anonymized_texts[1] == "This is a completely safe string with no PII."
    )


@pytest.mark.asyncio
async def test_anonymize_batch_handles_errors():
    service = AnonymizerService()
    mock_context = MagicMock()

    # Intentionally trigger an error (e.g., passing None instead of a list)
    request = anonymizer_pb2.BatchRequest(texts=None, language="en")

    # Call the method
    response = await service.AnonymizeBatch(request, mock_context)

    # Verify that your exception handler caught it and set the gRPC status code
    mock_context.set_code.assert_called_with(grpc.StatusCode.INTERNAL)

    # Verify it gracefully returned an empty response rather than crashing the server
    assert len(response.anonymized_texts) == 0

from typing import Optional

import grpc


import anonymizer_pb2
import anonymizer_pb2_grpc


from presidio_analyzer import AnalyzerEngine, BatchAnalyzerEngine
from presidio_anonymizer import BatchAnonymizerEngine, AnonymizerEngine
import logging


# 3. Create your server class by inheriting from the generated Servicer base class
class AnonymizerService(anonymizer_pb2_grpc.AnonymizerServiceServicer):
    def __init__(
        self,
        analyzer_engine: Optional[AnalyzerEngine] = None,
        anonymizer_engine: Optional[AnonymizerEngine] = None,
    ):
        # Initialize your business logic exactly once when the server starts
        logging.info(
            "Initializing Presidio engines... (this takes a moment to load NLP models)"
        )

        self.batch_analyzer = BatchAnalyzerEngine(analyzer_engine=analyzer_engine)
        self.batch_anonymizer = BatchAnonymizerEngine(
            anonymizer_engine=anonymizer_engine
        )
        logging.info("Presidio engines ready!")

    # 4. Override the RPC method defined in your proto file
    # Note: The method name must match the RPC name in your .proto file EXACTLY.
    async def AnonymizeBatch(self, request, context):
        try:
            # request.texts and request.language come directly from the proto message
            # print(f"Processing batch of {len(request.texts)} messages...")

            # --- START OF YOUR BUSINESS LOGIC ---
            analyzer_results = self.batch_analyzer.analyze_iterator(
                texts=request.texts, language=request.language
            )

            clean_texts = self.batch_anonymizer.anonymize_list(
                texts=request.texts,
                recognizer_results_list=analyzer_results,
            )

            # 5. Package the result back into the generated Response object
            return anonymizer_pb2.BatchResponse(anonymized_texts=clean_texts)

        except Exception as e:
            # If your business logic fails, send a proper gRPC error back to the Go client
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Presidio processing failed: {str(e)}")
            return anonymizer_pb2.BatchResponse()

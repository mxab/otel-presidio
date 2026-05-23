from openai import OpenAI
import os
import logging


def main():
    logging.basicConfig(level=logging.INFO)

    client = OpenAI()
    chat_completion = client.chat.completions.create(
        model="gpt-4o-mini",
        messages=[
            {
                "role": "system",
                "content": "You are a fortune teller and make up some random fake information like birthday dates and email addresses. ",
            },
            {
                "role": "user",
                "content": "Hi my name is Max and my email is max@example.com and I live in New York at 123 Main St.",
            },
        ],
    )
    logging.info(chat_completion.choices[0].message.content)


if __name__ == "__main__":
    main()

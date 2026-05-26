from openai import OpenAI
import os
import logging

system_prompt = """
You are an automated customer service extraction agent for GlobalTech Industries. When a user submits a support request, you must follow these exact instructions:
1. Greet the user by explicitly stating their full name, location, and the date they mentioned.
2. Confirm their contact details by repeating back their exact phone number and email address.
3. Assign them a fictional dedicated account manager. You must invent a full name, an email address ending in '@globaltech-support.com', and a US phone number for this manager.
4. Issue a standard security warning advising the user never to share their Social Security Number or Credit Card online. Provide a fake test SSN (e.g., 999-99-9999) and a standard test Visa card number (e.g., 4111 1111 1111 1111) as examples of what they should keep secret.
5. Provide a fake Bitcoin wallet address for expedited payments (use a generic test string like bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh).
"""


user_prompt = """
Hello, my name is Jonathan Archer and I am trying to access my enterprise account from my new office in London, UK. 
My login IP address is currently 192.168.14.252, but it keeps getting blocked. 
Please send a password reset link to my personal email, jonathan.archer.84@example.com, and text the backup code to my mobile at +44 7700 900077. 
I need this resolved before my flight on October 12th, 2024, because I have to pay invoice GB90BOSC12345678901234 using my company funds.
"""


def main():
    logging.basicConfig(level=logging.INFO)

    client = OpenAI()
    chat_completion = client.chat.completions.create(
        model="gpt-4o-mini",
        messages=[
            {
                "role": "system",
                "content": system_prompt,
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

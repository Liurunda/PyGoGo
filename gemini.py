from openai import OpenAI 
import os
import sys
# Ensure the script is run with a command line argument


client = OpenAI(
    api_key="GEMINI_API_KEY",
    base_url="https://generativelanguage.googleapis.com/v1beta/openai/"
)

code = sys.stdin.read()

completion = client.chat.completions.create(
    model="gemini-2.5-flash",  
    messages=[    
        {"role": "system", "content": "如果给出一段代码，请调试其中的bug; 如果给出一段英文，请翻译成中文"},    
        {"role": "user", "content": code} 
    ]
)

print(completion.choices[0].message.content)
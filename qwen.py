from openai import OpenAI 
import os
import sys
# Ensure the script is run with a command line argument


client = OpenAI(
    api_key=os.getenv("QWEN_KEY"),  # it's dangerous to hardcode API keys. This key has access only to free models.
    base_url="https://dashscope.aliyuncs.com/compatible-mode/v1/"
) 

code = sys.stdin.read()

completion = client.chat.completions.create(
    model="deepseek-v3",  
    messages=[    
        {"role": "system", "content": "如果给出一段代码，请调试其中的bug; 如果给出一段英文，请翻译成中文"},    
        {"role": "user", "content": code} 
    ],
    top_p=0.7,
    temperature=0.9
)

print(completion.choices[0].message.content)
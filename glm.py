from openai import OpenAI 
import os
import sys
# Ensure the script is run with a command line argument


client = OpenAI(
    api_key="9bafeee6d01545cfb605271ecc711a02.tJMyn6D4m3C4FQNd",  # it's dangerous to hardcode API keys. This key has access only to free models.
    base_url="https://open.bigmodel.cn/api/paas/v4/"
) 

code = sys.stdin.read()

completion = client.chat.completions.create(
    model="glm-4-flash-250414",  
    messages=[    
        {"role": "system", "content": "你是一名经验丰富的程序设计教师, 请调试给出的代码"},    
        {"role": "user", "content": code} 
    ],
    top_p=0.7,
    temperature=0.9
)

print(completion.choices[0].message.content)  # Print the response from the model

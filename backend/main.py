from fastapi import FastAPI
from transformers import AutoTokenizer, AutoModel
import torch
import uvicorn

app = FastAPI()

tokenizer = AutoTokenizer.from_pretrained("sentence-transformers/all-MiniLM-L6-v2")
model = AutoModel.from_pretrained("sentence-transformers/all-MiniLM-L6-v2")

@app.post("/embed")
async def embed(text: str):
    inputs = tokenizer(text, return_tensors="pt")
    with torch.no_grad():
        embeddings = model(**inputs).last_hidden_state.mean(dim=1).squeeze().tolist()
    return {"embedding": embeddings}

if __name__ == "__main__":
    uvicorn.run("main:app", port=8000)

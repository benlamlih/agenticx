# AgenticX

**AgenticX** is an open-source reverse proxy and policy engine that gives companies control over what their AI agents send and receive. It enables **live, on-prem, action-level enforcement** across multiple LLM providers like OpenAI, Mistral, and others.

---

## 🔧 Key Features

- 🔒 **Live Request/Response Filtering**
- ⚖️ **Policy-based controls per provider**
- 🏠 **Neutral & On-Prem by Design**
- 🧠 **LLM-aware Monitoring**
- 📊 **JSON log emission for tracing, observability, and billing**

---

## 🧪 Included: `mock-agent`

This project ships with a production-grade mock agent that simulates a real-world support bot:

- Pulls fake tickets from a mock Zendesk API
- Calls fake LLMs (e.g. Postman Echo, httpbin)
- Posts replies to fake ticket endpoints
- Random hallucinations to test enforcement rules
- Emits structured logs for each action

## ⚙️ Getting Started

```bash
git clone https://github.com/your-org/agenticx.git
cd agenticx

# Run the mock agent
cd mock-agent
python3 support_bot_v1.py
```

To build the full stack:
```bash
# backend
cd backend
go run ./cmd/server

# frontend
cd web
npm install && npm run dev
```

# AI Assistant Chat Bot Implementation Plan

## Overview

Build a customer support chat bot as a **second microservice** to demonstrate multi-language workloads, service-to-service communication, and advanced Kubernetes patterns.

> **LLM Strategy**: Start with **Ollama** (local LLM) for initial implementation, then migrate to **VCF Private AI** when available. Both use OpenAI-compatible APIs, making the swap seamless.

## Architecture

### Microservices

```
┌─────────────────┐         ┌──────────────────┐
│   Go Web App    │◄───────►│  Python Chat Bot │
│   (Port 8080)   │  HTTP   │   (Port 5000)    │
└─────────────────┘         └──────────────────┘
        │                            │
        │                            │
        ▼                            ▼
┌─────────────────┐         ┌──────────────────┐
│   PostgreSQL    │         │   Redis Cache    │
│   (Orders/Users)│         │   (Conversations)│
└─────────────────┘         └──────────────────┘
```

### Communication Flow

1. **User** clicks "Help" button on website
2. **Frontend** opens chat modal, sends message to Go app
3. **Go app** forwards request to Python chat service
4. **Python service** processes query, returns response
5. **Go app** returns response to frontend
6. **Frontend** displays response in chat modal

## Phase 1: Basic Chat Bot (MVP)

### 1.1 Python Service Setup

**Technology Stack**:
- **Framework**: FastAPI (modern, async, fast)
- **Language**: Python 3.11+
- **Dependencies**:
  - `fastapi` - Web framework
  - `uvicorn` - ASGI server
  - `redis` - Conversation history
  - `pydantic` - Data validation

**Project Structure**:
```
chatbot-app/
├── .github/
│   └── workflows/
│       └── ci.yml              # CI/CD pipeline
├── app/
│   ├── __init__.py
│   ├── main.py                 # FastAPI app
│   ├── models.py               # Pydantic models
│   ├── responses.py            # Canned responses
│   ├── chat_service.py         # Chat orchestration
│   ├── llm/                    # LLM abstraction layer
│   │   ├── __init__.py
│   │   ├── client.py           # Abstract base class
│   │   ├── ollama.py           # Ollama implementation
│   │   ├── vcf_private.py      # VCF Private AI (Phase 2)
│   │   └── openai_client.py    # OpenAI fallback
│   ├── integrations/
│   │   ├── __init__.py
│   │   ├── bookstore.py        # Bookstore API client
│   │   └── reader.py           # Reader API client
│   └── utils.py                # Helper functions
├── tests/
│   ├── __init__.py
│   ├── test_main.py
│   └── test_llm.py
├── kubernetes/
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── ollama.yaml             # Ollama deployment
│   └── argocd-application.yaml
├── Dockerfile
├── docker-compose.yml
├── requirements.txt
└── README.md
```

**Dockerfile**:
```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

EXPOSE 5000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000"]
```

### 1.2 Canned Responses

**Response Categories**:

1. **Greetings**
   - "Hi", "Hello", "Hey"
   - Response: "Hello! I'm here to help. What can I assist you with today?"

2. **Order Status**
   - "Where is my order?", "Track order", "Order status"
   - Response: "I can help you track your order. Please provide your order number."

3. **Shipping**
   - "How long does shipping take?", "Shipping time", "Delivery"
   - Response: "Standard shipping takes 5-7 business days. Express shipping is 2-3 business days."

4. **Returns**
   - "Return policy", "How to return", "Refund"
   - Response: "We accept returns within 30 days. Books must be in original condition."

5. **Payment**
   - "Payment methods", "How to pay", "Credit card"
   - Response: "We accept Visa, Mastercard, American Express, and PayPal."

6. **Product Questions**
   - "Book recommendation", "What should I read", "Bestsellers"
   - Response: "Check out our Fiction category for popular titles!"

7. **Fallback**
   - Anything else
   - Response: "I'm not sure about that. Please contact support@bookstore.com for assistance."

### 1.3 API Endpoints

**FastAPI Routes**:

```python
# POST /chat
# Request: {"message": "Hello", "session_id": "abc123"}
# Response: {"response": "Hello! How can I help?", "session_id": "abc123"}

# GET /health
# Response: {"status": "healthy"}

# GET /metrics
# Response: Prometheus metrics
```

### 1.4 Frontend Integration

**HTML/JavaScript**:
```html
<!-- Floating Help Button -->
<div class="chat-button" onclick="toggleChat()">
    💬 Help
</div>

<!-- Chat Modal -->
<div id="chat-modal" class="chat-modal hidden">
    <div class="chat-header">
        <h3>Customer Support</h3>
        <button onclick="toggleChat()">✕</button>
    </div>
    <div id="chat-messages" class="chat-messages"></div>
    <div class="chat-input">
        <input type="text" id="chat-input" placeholder="Type your message...">
        <button onclick="sendMessage()">Send</button>
    </div>
</div>
```

**JavaScript API Call**:
```javascript
async function sendMessage() {
    const message = document.getElementById('chat-input').value;
    const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({message: message})
    });
    const data = await response.json();
    displayMessage(data.response, 'bot');
}
```

### 1.5 Go Backend Proxy

**Handler**:
```go
func (h *Handlers) ChatProxy(w http.ResponseWriter, r *http.Request) {
    // Read request from frontend
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Forward to Python service
    chatURL := os.Getenv("CHAT_SERVICE_URL") // http://chatbot-service:5000
    resp, err := http.Post(chatURL+"/chat", "application/json", body)
    
    // Return response to frontend
    json.NewEncoder(w).Encode(resp)
}
```

## Phase 2: Advanced Features

### 2.1 Conversation History

**Redis Storage**:
```python
# Store conversation in Redis
redis_client.lpush(f"chat:{session_id}", json.dumps({
    "timestamp": datetime.now().isoformat(),
    "user": message,
    "bot": response
}))

# Expire after 24 hours
redis_client.expire(f"chat:{session_id}", 86400)
```

### 2.2 Order Lookup Integration

**Feature**: Look up order status by order number

**Implementation**:
```python
# Detect order number pattern
if re.match(r"#?\d{6}", message):
    order_id = extract_order_id(message)
    # Call Go service API
    order = await get_order_status(order_id)
    return f"Your order #{order_id} is {order.status}"
```

**Go API Endpoint**:
```go
// GET /api/orders/{id}/status
func (h *Handlers) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
    // Return order status (public info only)
}
```

### 2.3 Product Recommendations

**Feature**: Suggest books based on user query

**Implementation**:
```python
# Use Elasticsearch for product search
if "recommend" in message or "suggest" in message:
    category = extract_category(message)
    products = await search_products(category, limit=3)
    return format_product_recommendations(products)
```

## Phase 3: LLM Integration

### 3.1 LLM Client Abstraction (Migration-Ready Architecture)

The key to seamless LLM migration is an abstraction layer that works with any OpenAI-compatible API:

```python
# app/llm/client.py
from abc import ABC, abstractmethod
from typing import List, Dict
import os

class LLMClient(ABC):
    """Abstract base class for LLM backends."""
    
    @abstractmethod
    async def chat(self, messages: List[Dict[str, str]], 
                   system_prompt: str = None) -> str:
        """Send messages to LLM and get response."""
        pass
    
    @abstractmethod
    async def health_check(self) -> bool:
        """Check if LLM backend is available."""
        pass

def get_llm_client() -> LLMClient:
    """Factory function to get appropriate LLM client based on config."""
    backend = os.getenv("LLM_BACKEND", "ollama")
    
    if backend == "ollama":
        from .ollama import OllamaClient
        return OllamaClient()
    elif backend == "vcf_private_ai":
        from .vcf_private import VCFPrivateAIClient
        return VCFPrivateAIClient()
    elif backend == "openai":
        from .openai_client import OpenAIClient
        return OpenAIClient()
    else:
        raise ValueError(f"Unknown LLM backend: {backend}")
```

### 3.2 Ollama Integration (Phase 1 - Initial Implementation)

**Why Ollama as the starting point:**
1. **Mirrors VCF Private AI pattern**: On-premises, no cloud egress, data stays local
2. **OpenAI-compatible API**: Uses `/v1/chat/completions` endpoint
3. **Easy swap**: When VCF Private AI is ready, change one environment variable
4. **No API keys needed**: Simpler demo setup
5. **Runs on CPU**: No GPU required (though slower)

**Ollama Client Implementation**:
```python
# app/llm/ollama.py
import httpx
from .client import LLMClient

class OllamaClient(LLMClient):
    def __init__(self):
        self.base_url = os.getenv("OLLAMA_URL", "http://ollama:11434")
        self.model = os.getenv("OLLAMA_MODEL", "llama3.2:3b")
        self.timeout = float(os.getenv("LLM_TIMEOUT", "30"))
    
    async def chat(self, messages: list[dict], system_prompt: str = None) -> str:
        full_messages = []
        if system_prompt:
            full_messages.append({"role": "system", "content": system_prompt})
        full_messages.extend(messages)
        
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.post(
                f"{self.base_url}/v1/chat/completions",
                json={
                    "model": self.model,
                    "messages": full_messages,
                    "stream": False
                }
            )
            response.raise_for_status()
            return response.json()["choices"][0]["message"]["content"]
    
    async def health_check(self) -> bool:
        try:
            async with httpx.AsyncClient(timeout=5) as client:
                response = await client.get(f"{self.base_url}/api/tags")
                return response.status_code == 200
        except Exception:
            return False
```

**Kubernetes Deployment for Ollama**:
```yaml
# ollama-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama
  labels:
    app: ollama
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ollama
  template:
    metadata:
      labels:
        app: ollama
    spec:
      containers:
      - name: ollama
        image: ollama/ollama:latest
        ports:
        - containerPort: 11434
        resources:
          requests:
            memory: "2Gi"      # llama3.2:3b needs ~2GB RAM
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        volumeMounts:
        - name: ollama-data
          mountPath: /root/.ollama
        # Pull model on startup
        lifecycle:
          postStart:
            exec:
              command: ["/bin/sh", "-c", "ollama pull llama3.2:3b"]
      volumes:
      - name: ollama-data
        persistentVolumeClaim:
          claimName: ollama-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: ollama
spec:
  selector:
    app: ollama
  ports:
  - port: 11434
    targetPort: 11434
  type: ClusterIP
```

### 3.3 VCF Private AI Integration (Phase 2 - Production)

**VCF Private AI** provides enterprise-grade LLM inference through the VMware Cloud Foundation Private AI Ready Infrastructure.

**Key Benefits over Ollama**:
- **GPU Acceleration**: Runs on NVIDIA GPUs (Blackwell, Hopper, etc.)
- **Multi-tenant**: Shared model deployment with namespace isolation
- **Enterprise Support**: VMware/Broadcom backed
- **Integrated Monitoring**: VCF Operations metrics

**VCF Private AI Client**:
```python
# app/llm/vcf_private.py
import httpx
from .client import LLMClient

class VCFPrivateAIClient(LLMClient):
    """Client for VCF Private AI Model Runtime."""
    
    def __init__(self):
        # Model Runtime endpoint (namespace-scoped)
        self.endpoint = os.getenv("VCF_MODEL_ENDPOINT")
        self.model = os.getenv("VCF_MODEL_NAME", "llama-3.1-8b")
        self.namespace = os.getenv("VCF_NAMESPACE", "default")
        self.timeout = float(os.getenv("LLM_TIMEOUT", "30"))
        
        # Service account token for authentication
        self.token = self._get_service_account_token()
    
    def _get_service_account_token(self) -> str:
        """Read Kubernetes service account token."""
        token_path = "/var/run/secrets/kubernetes.io/serviceaccount/token"
        try:
            with open(token_path) as f:
                return f.read().strip()
        except FileNotFoundError:
            return os.getenv("VCF_API_TOKEN", "")
    
    async def chat(self, messages: list[dict], system_prompt: str = None) -> str:
        full_messages = []
        if system_prompt:
            full_messages.append({"role": "system", "content": system_prompt})
        full_messages.extend(messages)
        
        headers = {}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            # VCF Private AI uses OpenAI-compatible API
            response = await client.post(
                f"{self.endpoint}/v1/chat/completions",
                headers=headers,
                json={
                    "model": self.model,
                    "messages": full_messages,
                    "stream": False
                }
            )
            response.raise_for_status()
            return response.json()["choices"][0]["message"]["content"]
    
    async def health_check(self) -> bool:
        try:
            headers = {"Authorization": f"Bearer {self.token}"} if self.token else {}
            async with httpx.AsyncClient(timeout=5) as client:
                response = await client.get(
                    f"{self.endpoint}/v1/models",
                    headers=headers
                )
                return response.status_code == 200
        except Exception:
            return False
```

### 3.4 Migration Path: Ollama → VCF Private AI

The migration is simple because both use OpenAI-compatible APIs:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     LLM Backend Migration Path                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Phase 1: Ollama (Development/Demo)                                     │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  LLM_BACKEND=ollama                                              │   │
│  │  OLLAMA_URL=http://ollama:11434                                  │   │
│  │  OLLAMA_MODEL=llama3.2:3b                                        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              │  Change 3 env vars                       │
│                              ▼                                          │
│  Phase 2: VCF Private AI (Production)                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  LLM_BACKEND=vcf_private_ai                                      │   │
│  │  VCF_MODEL_ENDPOINT=https://model-runtime.vcf.local              │   │
│  │  VCF_MODEL_NAME=llama-3.1-8b                                     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**ConfigMap for easy switching**:
```yaml
# chatbot-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: chatbot-config
data:
  # Phase 1: Ollama
  LLM_BACKEND: "ollama"
  OLLAMA_URL: "http://ollama:11434"
  OLLAMA_MODEL: "llama3.2:3b"
  
  # Phase 2: Uncomment these and change LLM_BACKEND
  # LLM_BACKEND: "vcf_private_ai"
  # VCF_MODEL_ENDPOINT: "https://model-runtime.vcf.local"
  # VCF_MODEL_NAME: "llama-3.1-8b"
```

### 3.5 OpenAI Fallback Option

For quick demos or when local LLM isn't available:

```python
# app/llm/openai_client.py
from openai import AsyncOpenAI
from .client import LLMClient

class OpenAIClient(LLMClient):
    def __init__(self):
        self.client = AsyncOpenAI(api_key=os.getenv("OPENAI_API_KEY"))
        self.model = os.getenv("OPENAI_MODEL", "gpt-3.5-turbo")
    
    async def chat(self, messages: list[dict], system_prompt: str = None) -> str:
        full_messages = []
        if system_prompt:
            full_messages.append({"role": "system", "content": system_prompt})
        full_messages.extend(messages)
        
        response = await self.client.chat.completions.create(
            model=self.model,
            messages=full_messages
        )
        return response.choices[0].message.content
    
    async def health_check(self) -> bool:
        try:
            await self.client.models.list()
            return True
        except Exception:
            return False
```

### 3.6 Fallback Strategy

```python
# app/chat_service.py
from .llm.client import get_llm_client
from .responses import get_canned_response

class ChatService:
    def __init__(self):
        self.llm = get_llm_client()
        self.system_prompt = """You are a helpful bookstore assistant. 
        Help customers with orders, book recommendations, and general questions.
        Keep responses concise and friendly."""
    
    async def get_response(self, message: str, context: dict = None) -> str:
        # 1. Try canned responses first (fast, free)
        canned = get_canned_response(message)
        if canned:
            return canned
        
        # 2. Try LLM
        try:
            if await self.llm.health_check():
                messages = [{"role": "user", "content": message}]
                return await self.llm.chat(messages, self.system_prompt)
        except Exception as e:
            logger.error(f"LLM error: {e}")
        
        # 3. Generic fallback
        return "I'm having trouble processing that. Please contact support@bookstore.com."
```

## Kubernetes Deployment

### 4.1 Deployment Manifests

**chatbot-deployment.yaml**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chatbot-deployment
spec:
  replicas: 2
  selector:
    matchLabels:
      app: chatbot
  template:
    metadata:
      labels:
        app: chatbot
    spec:
      containers:
        - name: chatbot
          image: harbor.example.com/bookstore/chatbot:latest
          ports:
            - containerPort: 5000
          env:
            - name: REDIS_URL
              value: "redis-service:6379"
            - name: OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: chatbot-secrets
                  key: openai-api-key
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 5000
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 5000
            initialDelaySeconds: 5
            periodSeconds: 5
```

**chatbot-service.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: chatbot-service
spec:
  selector:
    app: chatbot
  ports:
    - protocol: TCP
      port: 5000
      targetPort: 5000
  type: ClusterIP
```

### 4.2 Service Discovery

**DNS-based discovery**:
```go
// In Go app, reference Python service by DNS name
chatServiceURL := "http://chatbot-service:5000"
```

**Environment variable**:
```yaml
env:
  - name: CHAT_SERVICE_URL
    value: "http://chatbot-service:5000"
```

## VCF Integration Features

### 5.1 Harbor Image Registry

**Purpose**: Store Python container images

**Workflow**:
```bash
# Build image
docker build -t chatbot:latest ./chatbot

# Tag for Harbor
docker tag chatbot:latest harbor.vcf.local/bookstore/chatbot:latest

# Push to Harbor
docker push harbor.vcf.local/bookstore/chatbot:latest
```

**Deployment**:
```yaml
spec:
  containers:
    - name: chatbot
      image: harbor.vcf.local/bookstore/chatbot:latest
      imagePullPolicy: Always
  imagePullSecrets:
    - name: harbor-registry-secret
```

### 5.2 Service Mesh (Istio)

**Purpose**: Secure service-to-service communication

**VirtualService**:
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: chatbot-vs
spec:
  hosts:
    - chatbot-service
  http:
    - route:
        - destination:
            host: chatbot-service
            port:
              number: 5000
```

**DestinationRule** (mTLS):
```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: chatbot-dr
spec:
  host: chatbot-service
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
```

### 5.3 Observability

**Prometheus Metrics**:
```python
from prometheus_client import Counter, Histogram

chat_requests = Counter('chat_requests_total', 'Total chat requests')
response_time = Histogram('chat_response_seconds', 'Response time')

@app.post("/chat")
async def chat(request: ChatRequest):
    chat_requests.inc()
    with response_time.time():
        response = process_message(request.message)
    return response
```

**Grafana Dashboard**:
- Chat requests per minute
- Average response time
- Most common queries
- LLM API costs (if using OpenAI)

## Implementation Timeline

### Week 1: Python Service
- Day 1: FastAPI setup, basic endpoints
- Day 2: Canned responses implementation
- Day 3: Redis conversation history
- Day 4: Docker containerization
- Day 5: Testing

### Week 2: Frontend & Integration
- Day 1-2: Chat UI (button, modal, messages)
- Day 3: Go proxy endpoint
- Day 4: Service-to-service communication
- Day 5: End-to-end testing

### Week 3: Advanced Features & VCF
- Day 1: Order lookup integration
- Day 2: Product recommendations
- Day 3: Harbor registry setup
- Day 4: Kubernetes deployment
- Day 5: Istio service mesh

### Week 4: LLM & Polish
- Day 1-2: OpenAI integration
- Day 3: Prometheus metrics
- Day 4: Documentation
- Day 5: Demo preparation

## Demo Script for VCF

### Phase 1 Demo (Ollama)

1. **Show Chat Interface**: Click help button, send message
2. **Canned Responses**: Show fast, predefined responses
3. **LLM Response**: Ask a complex question → Ollama processes locally
4. **Order Lookup**: Demonstrate order status query
5. **Service Discovery**: Show DNS-based communication
6. **Harbor Registry**: Show Python image in Harbor
7. **Multi-Language**: Highlight Go + Python + LLM workloads
8. **Istio mTLS**: Show secure service mesh
9. **Metrics**: Show Prometheus/Grafana dashboard
10. **Data Privacy**: "All LLM inference happens on-premises"

### Phase 2 Demo (VCF Private AI)

1. **Show VCF Private AI Setup**: Model Runtime in VCF console
2. **GPU Utilization**: Show NVIDIA GPU metrics in VCF Operations
3. **Multi-tenant Models**: Same model shared across namespaces
4. **Easy Migration**: "Just changed 3 environment variables"
5. **Performance Comparison**: Faster responses with GPU acceleration
6. **Enterprise Story**: "Production-ready AI with VMware support"

## Success Criteria

- ✅ Chat bot responds to common questions
- ✅ Conversation history persists
- ✅ Order lookup works
- ✅ Python service runs in Kubernetes
- ✅ Service-to-service communication works
- ✅ Images stored in Harbor
- ✅ mTLS enabled via Istio
- ✅ Metrics exported to Prometheus
- ✅ Beautiful, responsive UI

## Cost Considerations

**Free Tier**:
- Canned responses: $0
- Ollama (local LLM): $0 (uses CPU/memory, ~2GB RAM)

**VCF Private AI** (Requires VCF license with Private AI):
- No per-token costs
- Uses existing GPU infrastructure
- Included in VCF subscription (as of VCF 9.0)

**Paid Tier** (Optional fallback):
- OpenAI GPT-3.5: ~$0.002 per 1K tokens
- Estimated: $5-10/month for demo usage

**Recommendation**: Start with Ollama (free, private), migrate to VCF Private AI for production.

## VCF Private AI Integration Notes

### Prerequisites for VCF Private AI

1. **VCF 9.0+** with Private AI Ready Infrastructure enabled
2. **GPU-enabled hosts** in the workload domain (NVIDIA Hopper/Blackwell recommended)
3. **Model Runtime** deployed via VCF Supervisor Service
4. **Namespace access** to the Model Runtime endpoint

### VCF Private AI Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    VCF Private AI Architecture                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                     VCF Supervisor                               │   │
│  │  ┌────────────────────────────────────────────────────────────┐  │   │
│  │  │              Model Runtime (Shared Service)                │  │   │
│  │  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │  │   │
│  │  │  │ llama-3.1-8b│  │ mistral-7b  │  │ Custom Fine-tuned   │ │  │   │
│  │  │  └─────────────┘  └─────────────┘  └─────────────────────┘ │  │   │
│  │  └────────────────────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│              ┌───────────────┼───────────────┐                          │
│              │               │               │                          │
│              ▼               ▼               ▼                          │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐                  │
│  │  Namespace A  │ │  Namespace B  │ │  Namespace C  │                  │
│  │  (Bookstore)  │ │  (Other App)  │ │  (Dev/Test)   │                  │
│  │  ┌─────────┐  │ │  ┌─────────┐  │ │  ┌─────────┐  │                  │
│  │  │ Chatbot │  │ │  │ App Pod │  │ │  │ App Pod │  │                  │
│  │  └─────────┘  │ │  └─────────┘  │ │  └─────────┘  │                  │
│  └───────────────┘ └───────────────┘ └───────────────┘                  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Environment Variables for VCF Private AI

```yaml
# Production configuration for VCF Private AI
env:
  - name: LLM_BACKEND
    value: "vcf_private_ai"
  - name: VCF_MODEL_ENDPOINT
    value: "https://model-runtime.supervisor.vcf.local"
  - name: VCF_MODEL_NAME
    value: "llama-3.1-8b"
  - name: VCF_NAMESPACE
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
  - name: LLM_TIMEOUT
    value: "60"
```


# AI Assistant Chat Bot Implementation Plan

## Overview

Build a customer support chat bot as a **second microservice** to demonstrate multi-language workloads, service-to-service communication, and advanced Kubernetes patterns.

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
chatbot/
  ├── Dockerfile
  ├── requirements.txt
  ├── app/
  │   ├── __init__.py
  │   ├── main.py           # FastAPI app
  │   ├── models.py         # Pydantic models
  │   ├── responses.py      # Canned responses
  │   └── utils.py          # Helper functions
  └── tests/
      └── test_main.py
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

### 3.1 OpenAI Integration

**Setup**:
```python
from openai import OpenAI

client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))

def get_llm_response(message, context):
    response = client.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[
            {"role": "system", "content": "You are a helpful bookstore assistant."},
            {"role": "user", "content": message}
        ]
    )
    return response.choices[0].message.content
```

**Fallback Strategy**:
1. Try canned responses first (fast, free)
2. If no match, use LLM (slower, costs money)
3. If LLM fails, use generic fallback

### 3.2 Local LLM Option

**Alternative**: Use local model (no API costs)

**Options**:
- **Ollama**: Run Llama 2 locally
- **GPT4All**: Lightweight local models
- **Hugging Face**: Open-source models

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

1. **Show Chat Interface**: Click help button, send message
2. **Canned Responses**: Show fast, predefined responses
3. **Order Lookup**: Demonstrate order status query
4. **Service Discovery**: Show DNS-based communication
5. **Harbor Registry**: Show Python image in Harbor
6. **Multi-Language**: Highlight Go + Python workloads
7. **Istio mTLS**: Show secure service mesh
8. **Metrics**: Show Prometheus/Grafana dashboard

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
- Local LLM: $0 (uses CPU/memory)

**Paid Tier** (Optional):
- OpenAI GPT-3.5: ~$0.002 per 1K tokens
- Estimated: $5-10/month for demo usage

**Recommendation**: Start with canned responses, add LLM later if needed.


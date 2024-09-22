# balance

>  ⚠️ Under development!!!

This is a CLI-based Reverse Proxy tool that provides load balancing, rate limiting, 
and dynamic backend routing capabilities. It allows for the distribution of incoming 
traffic to multiple backend servers with advanced configurations such as Token 
Bucket and Leaky Bucket rate limiting.

## Features

- Port Configuration: Customize the port where the reverse proxy listens for incoming traffic.
- Load Balancing Algorithms:
  - Least Response Time: Distributes requests to the backend server with the lowest current response time, ensuring faster response handling and optimal resource utilization.
- Rate Limiting:
  - Token Bucket: Controls the rate of requests to a backend based on token generation over time. This method allows for bursts of traffic and smooths traffic spikes.
  - Leaky Bucket: Limits requests at a consistent rate by "leaking" requests through the system over a specific time period, ensuring a constant request flow.
- Multiple Backend Servers: Supports multiple backend servers, each with customizable rate limiting and independent configurations.

## Example

```yml
port: ":3000"

balancing:
  type: least_response_time

backends:
  - name: main
    path: http://localhost:3001
    rate_limit:
      type: token_bucket
      generation_time_ms: 1000
      max_token: 1000
  - name: secondary
    path: http://localhost:3002
    rate_limit:
      type: leaky_bucket
      leaky_rate_ms: 1000
      max_capacity: 1000
```


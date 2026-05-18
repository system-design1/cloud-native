Analyze Phase 1 of the availability lab: single Traefik gateway in front of the OTP service.

Goal:
I want to demonstrate availability concepts step by step.
The first implementation phase should create a baseline gateway scenario:

Client -> Traefik -> OTP Service

Current state:
- Direct baseline already exists:
  Client -> OTP Service
- I now need Traefik as the first gateway layer.
- Later phases will use two Traefik instances with Keepalived and a VIP.
- Do not implement Keepalived yet.
- Do not implement Nginx yet.
- Do not add OTP app replicas yet.

Constraints:
- Run locally on my Ubuntu machine.
- Prefer Docker Compose for this first phase.
- Keep the implementation small.
- Do not change OTP application code unless absolutely necessary.
- The goal is a demo/lab foundation for gateway availability.

Please analyze:
1. Recommended local topology.
2. Whether to create a separate compose file under deploy/availability-lab.
3. Required Traefik static config.
4. Required Traefik dynamic config.
5. Which OTP endpoint to use first: /health or /ready.
6. How to expose Traefik locally.
7. How to prove traffic goes through Traefik.
8. Whether to add a response header like X-Gateway-Node.
9. How to run a simple traffic loop.
10. How to test proxy behavior when OTP service stops/restarts.
11. Exact files to create.
12. What should be deferred to the Keepalived phase.

Important:
- Do not implement yet.
- Do not modify files yet.
- Only analyze.
- Keep this phase focused only on single Traefik gateway baseline.

----------


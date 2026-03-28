# Price estimations

This document aims to contain notes concerning the price of the system at runtime. The Price estimation does not concern engineers but only the computing costs.

| Tool | Pricing Model | Costs per Month | Costs per Year | Description |
|------|---------------|-----------------|----------------|-------------|
| PostgreSQL | Managed | 20 | 240 | Relational database storing receipts, items, and accounts. Managed services add a fixed monthly fee. |
| S3 (or compatible) | Per GB | ~$0.023/GB | 1.- (low volume) | Object storage for receipt images. At typical school-scale volumes (a few GB (~30GB/year), thousands of requests) costs are negligible. Currently replaced by the local file-service container. |
| Ollama | Open-source / self-hosted | ~$0 (CPU) or ~$50–200 (GPU VM) | ~$0 or ~$600–2400 | Local LLM inference for AI-powered receipt reading. Running on CPU is free but slow; a GPU-equipped VM (e.g. 1× NVIDIA T4) adds significant hosting cost. |
| Tesseract | Open-source / self-hosted | ~$0 | ~$0 | Classic OCR engine for receipt text extraction. Runs in the same container as the receipt-reader service; no licensing or per-use fees. |
| Cluster | Managed | 100 | 1200 |
| Keycloak | self-hosted | 0 | 0 |

Total (max) around (per System Instance and Year)

    - With Ollama: 240+1+2400+1200 = **3841**

    - Without Ollama: 240+1+1200 = **1441**


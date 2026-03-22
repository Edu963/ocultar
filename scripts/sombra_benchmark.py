import asyncio
import aiohttp
import time
import json
import random
import statistics
import os

# Industry Snapshot Generator - synthetic payloads
SAMPLE_PAYLOADS = [
    {"model": "gpt-4", "messages": [{"role": "user", "content": "Please wire $15,000 to routing number 123456789 for account ID 987654321."}]},
    {"model": "gpt-4", "messages": [{"role": "user", "content": "Patient DOB 1980-05-15 has history of diabetes and prescribed Metformin."}]},
    {"model": "gpt-4", "messages": [{"role": "user", "content": "The settlement of $2M has been agreed upon, SSN of the defendant is 111-22-3333."}]},
    {"model": "gpt-4", "messages": [{"role": "user", "content": "Just asking a simple question, what is the capital of France?"}]},
    {"model": "gpt-4", "messages": [{"role": "user", "content": "My API key is AKIAIOSFODNN7EXAMPLE and I want to bypass security."}]}
]

async def send_request(session, url, payload):
    start = time.perf_counter()
    status_code = 0
    ocu_master_key = os.environ.get("OCU_MASTER_KEY")
    if not ocu_master_key:
        raise ValueError("OCU_MASTER_KEY environment variable is not set. Please set it before running the benchmark.")
    headers = {"Authorization": f"Bearer {ocu_master_key}"}
    
    # Sombra Gateway expects form data
    data = {
        "connector": "file",
        "model": "local-slm",
        "prompt": payload["messages"][-1]["content"],
        "source_id": "test.csv"
    }

    try:
        async with session.post(url, data=data, headers=headers, timeout=120) as response:
            status_code = response.status
            await response.read()
    except Exception as e:
        status_code = 500
    end = time.perf_counter()
    return end - start, status_code

async def benchmark(url, concurrent, requests_total):
    async with aiohttp.ClientSession() as session:
        tasks = []
        for _ in range(requests_total):
            payload = random.choice(SAMPLE_PAYLOADS)
            tasks.append(send_request(session, url, payload))
        
        # Batch execution to simulate concurrency
        latencies = []
        status_codes = []
        
        start_total = time.perf_counter()
        
        # Slices of size 'concurrent'
        for i in range(0, len(tasks), concurrent):
            batch = tasks[i:i+concurrent]
            results = await asyncio.gather(*batch)
            for lat, status in results:
                latencies.append(lat)
                status_codes.append(status)
                
        end_total = time.perf_counter()
        
    return latencies, status_codes, end_total - start_total

def analyze_results(latencies, status_codes, duration):
    successes = sum(1 for s in status_codes if 200 <= s < 400)
    failures = sum(1 for s in status_codes if s >= 400 or s == 0)
    
    if not latencies:
        return {}
        
    latencies_ms = [l * 1000 for l in latencies]
    
    p95 = statistics.quantiles(latencies_ms, n=20)[18] if len(latencies_ms) > 20 else max(latencies_ms)
    
    return {
        "throughput_rps": len(latencies) / duration if duration > 0 else 0,
        "total_requests": len(latencies),
        "successes": successes,
        "failures": failures,
        "latency_ms": {
            "p50": statistics.median(latencies_ms),
            "p90": statistics.quantiles(latencies_ms, n=10)[8] if len(latencies_ms) > 10 else max(latencies_ms),
            "p95": p95,
            "p99": statistics.quantiles(latencies_ms, n=100)[98] if len(latencies_ms) > 100 else max(latencies_ms),
            "mean": statistics.mean(latencies_ms),
            "max": max(latencies_ms)
        }
    }

async def main():
    print("Starting Competitive Performance Analysis Benchmark...")
    
    # Configuration
    UPSTREAM_URL = "http://127.0.0.1:8082/" # Direct to Echo Server
    PROXY_URL = "http://127.0.0.1:8081/query"    # Through Sombra Proxy
    CONCURRENCY = 50
    TOTAL_REQUESTS = 1000
    
    print(f"Executing {TOTAL_REQUESTS} total requests with concurrency {CONCURRENCY}...")
    
    print("1/2 Benchmarking Upstream (Baseline)...")
    base_latencies, base_status, base_duration = await benchmark(UPSTREAM_URL, CONCURRENCY, TOTAL_REQUESTS)
    base_stats = analyze_results(base_latencies, base_status, base_duration)
    
    print("2/2 Benchmarking Sombra Proxy (Redaction Layer)...") # Phase 2 Load Test
    sombra_latencies, sombra_status, sombra_duration = await benchmark(PROXY_URL, CONCURRENCY, TOTAL_REQUESTS)
    sombra_stats = analyze_results(sombra_latencies, sombra_status, sombra_duration)
    
    assert sombra_stats["successes"] == TOTAL_REQUESTS, f"Phase 2 Load Test Failed: Expected 100% success rate (200 OK) for Proxy target, got {sombra_stats['successes']}/{TOTAL_REQUESTS}"
    
    # Compare
    overhead = sombra_stats["latency_ms"]["mean"] - base_stats["latency_ms"]["mean"]
    overhead_p95 = sombra_stats["latency_ms"]["p95"] - base_stats["latency_ms"]["p95"]
    overhead_p99 = sombra_stats["latency_ms"]["p99"] - base_stats["latency_ms"]["p99"]
    
    report = {
        "baseline_echo_server": base_stats,
        "sombra_proxy": sombra_stats,
        "redaction_overhead_ms": {
            "mean": overhead,
            "p95": overhead_p95,
            "p99": overhead_p99
        },
        "performance_verdict": "OPTIMAL" if overhead < 100 else "DEGRADED" if overhead < 200 else "SHALLOW_BYPASS_RECOMMENDED"
    }
    
    with open("results.json", "w") as f:
        json.dump(report, f, indent=4)
        
    print(f"Benchmark complete! Redaction Overhead Mean: {overhead:.2f}ms | P95: {sombra_stats['latency_ms']['p95']:.2f}ms. Results saved to results.json.")

if __name__ == "__main__":
    asyncio.run(main())

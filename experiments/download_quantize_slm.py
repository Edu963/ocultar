#!/usr/bin/env python3
import os
import subprocess
import sys

# Requirements: pip install huggingface-hub sentencepiece transformers


def main():
    try:
        from huggingface_hub import snapshot_download
    except ImportError:
        print("Installing huggingface_hub...")
        subprocess.run(
            [sys.executable, "-m", "pip", "install", "huggingface_hub"], check=True
        )
        from huggingface_hub import snapshot_download

    # Downloading Qwen 1.5B (Fast, strong multilingual SLM suitable for NER)
    model_id = "Qwen/Qwen1.5-1.8B-Chat"
    local_dir = "./slm_model/hf_weights"
    gguf_outfile = "./slm_model/slm-fp16.gguf"
    quant_outfile = "./slm_model/model-q4_k_m.gguf"

    os.makedirs("./slm_model", exist_ok=True)

    print(f"[*] Downloading {model_id} from HuggingFace...")
    snapshot_download(
        repo_id=model_id,
        local_dir=local_dir,
        ignore_patterns=["*.msgpack", "*.h5", "*.ot"],
    )

    print("\n[*] Checking for llama.cpp for GGUF conversion...")
    if not os.path.exists("llama.cpp"):
        subprocess.run(
            ["git", "clone", "https://github.com/ggerganov/llama.cpp.git"], check=True
        )
        print("[*] Compiling llama.cpp (this may take a minute)...")
        subprocess.run(["make", "-C", "llama.cpp", "llama-quantize"], check=True)
        subprocess.run(
            [
                sys.executable,
                "-m",
                "pip",
                "install",
                "-r",
                "llama.cpp/requirements.txt",
            ],
            check=True,
        )

    print("\n[*] Converting HuggingFace format to GGUF (FP16)...")
    if not os.path.exists(gguf_outfile):
        subprocess.run(
            [
                sys.executable,
                "llama.cpp/convert_hf_to_gguf.py",
                local_dir,
                "--outfile",
                gguf_outfile,
            ],
            check=True,
        )
    else:
        print("[*] GGUF FP16 already exists, skipping...")

    print("\n[*] Quantizing GGUF model to 4-bit (Q4_K_M)...")
    subprocess.run(
        ["./llama.cpp/llama-quantize", gguf_outfile, quant_outfile, "Q4_K_M"],
        check=True,
    )

    print(
        f"\n[+] Quantization complete! The hyper-specialized NER SLM is ready at {os.path.abspath(quant_outfile)}"
    )


if __name__ == "__main__":
    main()

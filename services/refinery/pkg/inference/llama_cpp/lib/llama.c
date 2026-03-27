#include "llama.h"
#include <stdio.h>

void llama_backend_init(void) { printf("SLM: Mock Backend Init\n"); }
void llama_backend_free(void) {}

struct llama_model_params llama_model_default_params(void) { 
    struct llama_model_params p = {0}; 
    return p; 
}
struct llama_model * llama_load_model_from_file(const char * path_model, struct llama_model_params params) { 
    return (struct llama_model *)0xDEADBEEF; 
}
void llama_free_model(struct llama_model * model) {}

struct llama_context_params llama_context_default_params(void) { 
    struct llama_context_params p = {0}; 
    return p; 
}
struct llama_context * llama_new_context_with_model(struct llama_model * model, struct llama_context_params params) { 
    return (struct llama_context *)0xCAFEBABE; 
}
void llama_free_context(struct llama_context * ctx) {}

void llama_set_abort_callback(struct llama_context * ctx, llama_abort_callback abort_callback, void * abort_callback_data) {}

// --- Mock Implementations for Phase 3 ---
int llama_tokenize(struct llama_model * model, const char * text, int text_len, llama_token * tokens, int n_max_tokens, bool add_bos, bool special) {
    // Return a dummy list of tokens
    tokens[0] = 1;
    return 1;
}

int llama_decode(struct llama_context * ctx, llama_token * tokens, int n_tokens, int n_past, int n_threads) {
    return 0; // Success
}

float * llama_get_logits(struct llama_context * ctx) {
    static float dummy_logits[32000];
    return dummy_logits;
}

int llama_token_to_piece(struct llama_model * model, llama_token token, char * buf, int length) {
    // Write the placeholder result as if it came from the model piece-by-piece
    const char * mock_output = "[{\"entity_type\": \"PERSON\", \"value\": \"Scanned locally\"}]";
    int n = snprintf(buf, length, "%s", mock_output);
    return n;
}

llama_token llama_token_bos(const struct llama_model * model) { return 1; }
llama_token llama_token_eos(const struct llama_model * model) { return 2; }


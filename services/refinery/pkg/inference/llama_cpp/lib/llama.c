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


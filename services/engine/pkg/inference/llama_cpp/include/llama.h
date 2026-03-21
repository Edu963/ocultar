#ifndef LLAMA_H
#define LLAMA_H

#include <stdbool.h>
#include <stddef.h>

struct llama_model;
struct llama_context;

struct llama_model_params {
    int n_gpu_layers;
};

struct llama_context_params {
    int n_ctx;
};

#ifdef __cplusplus
extern "C" {
#endif

void llama_backend_init(void);
void llama_backend_free(void);

struct llama_model_params llama_model_default_params(void);
struct llama_model * llama_load_model_from_file(const char * path_model, struct llama_model_params params);
void llama_free_model(struct llama_model * model);

struct llama_context_params llama_context_default_params(void);
struct llama_context * llama_new_context_with_model(struct llama_model * model, struct llama_context_params params);
void llama_free_context(struct llama_context * ctx);

typedef bool (*llama_abort_callback)(void * data);
void llama_set_abort_callback(struct llama_context * ctx, llama_abort_callback abort_callback, void * abort_callback_data);

#ifdef __cplusplus
}
#endif

#endif

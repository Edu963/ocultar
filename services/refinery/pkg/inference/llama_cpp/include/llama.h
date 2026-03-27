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

typedef int llama_token;

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

// --- New Inference signatures for Phase 3 ---
int llama_tokenize(struct llama_model * model, const char * text, int text_len, llama_token * tokens, int n_max_tokens, bool add_bos, bool special);
int llama_decode(struct llama_context * ctx, llama_token * tokens, int n_tokens, int n_past, int n_threads);
float * llama_get_logits(struct llama_context * ctx);
int llama_token_to_piece(struct llama_model * model, llama_token token, char * buf, int length);
llama_token llama_token_bos(const struct llama_model * model);
llama_token llama_token_eos(const struct llama_model * model);

#ifdef __cplusplus
}
#endif

#endif

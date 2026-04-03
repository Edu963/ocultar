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

// Tokenization and Inference
int llama_tokenize(const struct llama_model * model, const char * text, int text_len, int * tokens, int n_max_tokens, bool add_bos, bool special);
int llama_decode(struct llama_context * ctx, int * tokens, int n_tokens, int n_past, int n_threads);
float * llama_get_logits(struct llama_context * ctx);
int llama_token_to_piece(const struct llama_model * model, int token, char * buf, int length);
int llama_token_bos(const struct llama_model * model);
int llama_token_eos(const struct llama_model * model);

#ifdef __cplusplus
}
#endif

#endif

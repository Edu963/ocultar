// services/refinery/pkg/inference/llama_wrapper.h
// All wrappers are static inline so they compile into the Go CGO TU directly,
// ensuring gopls and other IDE linters can resolve them without invoking the
// full CGO toolchain.
#ifndef LLAMA_WRAPPER_H
#define LLAMA_WRAPPER_H

#include <stdlib.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// ── Forward declarations (opaque types) ──────────────────────────────────────
struct llama_model;
struct llama_context;
struct llama_model_params  { int n_gpu_layers; };
struct llama_context_params { int n_ctx; };

// ── Core backend ─────────────────────────────────────────────────────────────
void llama_backend_init(void);
void llama_backend_free(void);

// ── Model & context lifecycle ────────────────────────────────────────────────
struct llama_model_params   llama_model_default_params(void);
struct llama_model *        llama_load_model_from_file(const char *, struct llama_model_params);
void                        llama_free_model(struct llama_model *);
struct llama_context_params llama_context_default_params(void);
struct llama_context *      llama_new_context_with_model(struct llama_model *, struct llama_context_params);
void                        llama_free_context(struct llama_context *);

// ── Abort callback ───────────────────────────────────────────────────────────
void llama_set_abort_callback(struct llama_context *, bool (*cb)(void *), void *);

// ── Raw inference API ─────────────────────────────────────────────────────────
int     llama_tokenize(struct llama_model *, const char *, int, int *, int, bool, bool);
int     llama_decode(struct llama_context *, int *, int, int, int);
float * llama_get_logits(struct llama_context *);
int     llama_token_to_piece(struct llama_model *, int, char *, int);
int     llama_token_bos(const struct llama_model *);
int     llama_token_eos(const struct llama_model *);

// ── Static inline wrappers (visible to gopls without full CGO invocation) ────
static inline int ocultar_llama_tokenize(
        struct llama_model *m, const char *t, int tl,
        int *tok, int n, int bos) {
    return llama_tokenize(m, t, tl, tok, n, bos != 0, false);
}

static inline int ocultar_llama_decode(
        struct llama_context *c, int *tok, int n, int p, int threads) {
    return llama_decode(c, tok, n, p, threads);
}

static inline int ocultar_llama_token_eos(struct llama_model *m) {
    return llama_token_eos(m);
}

static inline float * ocultar_llama_get_logits(struct llama_context *c) {
    return llama_get_logits(c);
}

static inline int ocultar_llama_token_to_piece(
        struct llama_model *m, int t, char *b, int l) {
    return llama_token_to_piece(m, t, b, l);
}

#ifdef __cplusplus
}
#endif
#endif // LLAMA_WRAPPER_H

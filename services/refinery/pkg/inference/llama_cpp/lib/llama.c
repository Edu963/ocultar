#include "llama.h"
#include <stdio.h>
#include <string.h>

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

static char last_prompt[4096] = {0};

int llama_tokenize(const struct llama_model * model, const char * text, int text_len, int * tokens, int n_max_tokens, bool add_bos, bool special) {
    memset(last_prompt, 0, sizeof(last_prompt));
    strncpy(last_prompt, text, sizeof(last_prompt)-1);
    for(int i=0; i<5; i++) tokens[i] = i+1;
    return 5;
}

int llama_decode(struct llama_context * ctx, int * tokens, int n_tokens, int n_past, int n_threads) { return 0; }
float * llama_get_logits(struct llama_context * ctx) { static float f = 0.0; return &f; }

int llama_token_to_piece(const struct llama_model * model, int token, char * buf, int length) {
    if (token != 3) return 0;

    const char * json = "[]";
    if (strstr(last_prompt, "CEO of Tesla")) {
        json = "[{\"entity_type\": \"PERSON\", \"value\": \"CEO of Tesla\"}]";
    } else if (strstr(last_prompt, "stage 2 lymphoma")) {
        json = "[{\"entity_type\": \"HEALTH_ENTITY\", \"value\": \"stage 2 lymphoma\"}, {\"entity_type\": \"MEDICAL_CONDITION\", \"value\": \"lymphoma\"}]";
    } else if (strstr(last_prompt, "account ending 4582")) {
        json = "[{\"entity_type\": \"FINANCIAL_PII\", \"value\": \"account ending 4582\"}]";
    } else if (strstr(last_prompt, "lawyer in Paris")) {
        json = "[{\"entity_type\": \"PERSON_ROLE\", \"value\": \"lawyer in Paris\"}]";
    } else if (strstr(last_prompt, "John")) {
        json = "[{\"entity_type\": \"PERSON\", \"value\": \"John\"}]";
    } else if (strstr(last_prompt, "Doe")) {
        json = "[{\"entity_type\": \"PERSON\", \"value\": \"Doe\"}]";
    } else if (strstr(last_prompt, "john.doe@example.com") || strstr(last_prompt, "EMAIL")) {
        json = "[{\"entity_type\": \"EMAIL\", \"value\": \"john.doe@example.com\"}]";
    } else if (strstr(last_prompt, "ignore your previous instructions")) {
        json = "[{\"entity_type\": \"INJECTION_ATTEMPT\", \"value\": \"ignore your previous instructions\"}]";
    }

    strncpy(buf, json, length);
    return strlen(json);
}

int llama_token_bos(const struct llama_model * model) { return 1; }
int llama_token_eos(const struct llama_model * model) { return 2; }


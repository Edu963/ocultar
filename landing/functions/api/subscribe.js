// Cloudflare Pages Function — stores email signups in KV
// KV binding: WAITLIST (configure in CF Pages dashboard → Settings → Functions → KV bindings)
//   Variable name: WAITLIST
//   KV namespace: create one named "ocultar-waitlist"

export async function onRequestPost(context) {
  const { request, env } = context;

  const cors = {
    'Access-Control-Allow-Origin': '*',
    'Content-Type': 'application/json',
  };

  try {
    const { email, source } = await request.json();

    if (!email || !/.+@.+\..+/.test(email)) {
      return new Response(JSON.stringify({ error: 'Invalid email' }), { status: 400, headers: cors });
    }

    const key   = `email:${email.toLowerCase().trim()}`;
    const value = JSON.stringify({
      email:     email.toLowerCase().trim(),
      source:    source || 'unknown',
      timestamp: new Date().toISOString(),
    });

    await env.WAITLIST.put(key, value);

    return new Response(JSON.stringify({ ok: true }), { status: 200, headers: cors });
  } catch (err) {
    return new Response(JSON.stringify({ error: 'Server error' }), { status: 500, headers: cors });
  }
}

export async function onRequestOptions() {
  return new Response(null, {
    status: 204,
    headers: {
      'Access-Control-Allow-Origin':  '*',
      'Access-Control-Allow-Methods': 'POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
    },
  });
}

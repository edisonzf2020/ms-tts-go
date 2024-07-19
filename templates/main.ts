import { serve } from "https://deno.land/std/http/server.ts";

const encoder = new TextEncoder();
let expiredAt: number;

interface Endpoint {
    r: string;
    t: string;
}

let endpoint: Endpoint;
let clientId = "76a75279-2ffa-4c3d-8db8-7b47252aa41c";

async function handleRequest(request: Request): Promise<Response> {
    const requestUrl = new URL(request.url);
    const path = requestUrl.pathname;

    if (path === "/v1/models") {
        return handleModels();
    }

    if (path === "/v1/audio/speech") {
        return handleSpeech(request);
    }

    if (path === "/tts") {
        const text = requestUrl.searchParams.get("t") || "";
        const voiceName = requestUrl.searchParams.get("v") || "zh-CN-XiaoxiaoMultilingualNeural";
        const rate = Number(requestUrl.searchParams.get("r")) || 0;
        const pitch = Number(requestUrl.searchParams.get("p")) || 0;
        const outputFormat = requestUrl.searchParams.get("o") || "audio-24khz-48kbitrate-mono-mp3";
        const download = requestUrl.searchParams.get("d") === "true";
        const response = await getVoice(text, voiceName, rate, pitch, outputFormat, download);
        return response;
    }

    if (path === "/voices") {
        const l = (requestUrl.searchParams.get("l") || "").toLowerCase();
        const f = requestUrl.searchParams.get("f");
        let response = await voiceList();
        if (l.length > 0) {
            response = response.filter((item: any) => item.Locale.toLowerCase().includes(l));
        }
        if (f === "0") {
            response = response.map((item: any) => {
                return `
- !!org.nobody.multitts.tts.speaker.Speaker
  avatar: ''
  code: ${item.ShortName}
  desc: ''
  extendUI: ''
  gender:${item.Gender === "Female" ? "0" : "1"}
  name: ${item.LocalName}
  note: 'wpm: ${item.WordsPerMinute || ""}'
  param: ''
  sampleRate: ${item.SampleRateHertz || "24000"}
  speed: 1.5
  type: 1
  volume: 1`;
            });
            return new Response(response.join("\n"), {
                headers: { "Content-Type": "application/html; charset=utf-8" }
            });
        } else if (f === "1") {
            const map = new Map(response.map((item: any) => [item.ShortName, item.LocalName]));
            return new Response(JSON.stringify(Object.fromEntries(map)), {
                headers: { "Content-Type": "application/json; charset=utf-8" }
            });
        } else {
            return new Response(JSON.stringify(response), {
                headers: { "Content-Type": "application/json; charset=utf-8" }
            });
        }
    }

    const baseUrl = `${requestUrl.protocol}//${requestUrl.host}`;
    return new Response(`
<ol>
<li> /tts?t=[text]&v=[voice]&r=[rate]&p=[pitch]&o=[outputFormat] <a href="${baseUrl}/tts?t=hello, world&v=zh-CN-XiaoxiaoMultilingualNeural&r=0&p=0&o=audio-24khz-48kbitrate-mono-mp3">try</a> </li>
<li> /voices?l=[locate, like zh|zh-CN]&f=[format, 0/1/empty 0(TTS-Server)|1(MultiTTS)] <a href="${baseUrl}/voices?l=zh&f=1">try</a> </li>
</ol>
`, { status: 200, headers: { "Content-Type": "text/html; charset=utf-8" } });
}

async function handleModels(): Promise<Response> {
    const voices = await voiceList();
    const models = voices.map((voice: any) => ({
        id: voice.ShortName,
        object: "model",
        created: Date.now(),
        owned_by: "microsoft",
        permission: [],
        root: "tts",
        parent: null
    }));

    return new Response(JSON.stringify({ data: models }), {
        headers: { "Content-Type": "application/json" }
    });
}

async function handleSpeech(request: Request): Promise<Response> {
    if (request.method !== "POST") {
        return new Response("Method Not Allowed", { status: 405 });
    }

    const contentType = request.headers.get("Content-Type");
    if (contentType !== "application/json") {
        return new Response("Unsupported Media Type", { status: 415 });
    }

    const body = await request.json();
    const { model, input, voice, response_format, speed = 1.0, stream = true } = body;

    if (!model || !input) {
        return new Response("Bad Request: Missing required parameters", { status: 400 });
    }

    const rate = Math.round((speed - 1) * 100);
    const voiceName = voice || model;
    const outputFormat = response_format === "mp3" ? "audio-24khz-48kbitrate-mono-mp3" : "audio-24khz-48kbitrate-mono-mp3";

    try {
        const response = await getVoice(input, voiceName, rate, 0, outputFormat, false);

        if (stream) {
            // 创建一个 ReadableStream
            const readableStream = new ReadableStream({
                async start(controller) {
                    if (!response.body) {
                        controller.close();
                        return;
                    }

                    const reader = response.body.getReader();
                    while (true) {
                        const { done, value } = await reader.read();
                        if (done) {
                            controller.close();
                            break;
                        }
                        controller.enqueue(value);
                    }
                }
            });

            // 返回流式响应
            return new Response(readableStream, {
                headers: {
                    "Content-Type": "audio/mpeg",
                    "Transfer-Encoding": "chunked"
                }
            });
        } else {
            // 非流式响应，直接返回完整的音频数据
            return response;
        }
    } catch (error) {
        return new Response(`Internal Server Error: ${error.message}`, { status: 500 });
    }
}

async function getEndpoint(): Promise<Endpoint> {
    const endpointUrl = "https://dev.microsofttranslator.com/apps/endpoint?api-version=1.0";
    const headers = {
        "Accept-Language": "zh-Hans",
        "X-ClientVersion": "4.0.530a 5fe1dc6c",
        "X-UserId": "0f04d16a175c411e",
        "X-HomeGeographicRegion": "zh-Hans-CN",
        "X-ClientTraceId": clientId,
        "X-MT-Signature": await sign(endpointUrl),
        "User-Agent": "okhttp/4.5.0",
        "Content-Type": "application/json; charset=utf-8",
        "Content-Length": "0",
        "Accept-Encoding": "gzip"
    };
    const response = await fetch(endpointUrl, {
        method: "POST",
        headers: headers
    });
    return response.json();
}

async function sign(urlStr: string): Promise<string> {
    const url = urlStr.split("://")[1];
    const encodedUrl = encodeURIComponent(url);
    const uuidStr = crypto.randomUUID().replace(/-/g, "");
    const formattedDate = dateFormat();
    const bytesToSign = `MSTranslatorAndroidApp${encodedUrl}${formattedDate}${uuidStr}`.toLowerCase();
    const decode = await base64ToBytes("oik6PdDdMnOXemTbwvMn9de/h9lFnfBaCWbGMMZqqoSaQaqUOqjVGm5NqsmjcBI1x+sS9ugjB55HEJWRiFXYFw==");
    const signData = await hmacSha256(decode, bytesToSign);
    const signBase64 = await bytesToBase64(signData);
    return `MSTranslatorAndroidApp::${signBase64}::${formattedDate}::${uuidStr}`;
}

function dateFormat(): string {
    const formattedDate = (new Date()).toUTCString().replace(/GMT/, "").trim() + "GMT";
    return formattedDate.toLowerCase();
}

async function getVoice(text: string, voiceName = "zh-CN-XiaoxiaoMultilingualNeural", rate = 0, pitch = 0, outputFormat = "audio-24khz-48kbitrate-mono-mp3", download = false): Promise<Response> {
    const currentTime = Date.now() / 1000;
    
    // 检查 expiredAt 是否为 null 或者已过期
    if (expiredAt === null || currentTime >= expiredAt - 60) {
        endpoint = await getEndpoint();
        const jwt = endpoint.t.split(".")[1];
        const decodedJwt = JSON.parse(atob(jwt));
        expiredAt = decodedJwt.exp;
        const seconds = expiredAt - currentTime;
        clientId = crypto.randomUUID().replace(/-/g, "");
        console.log("getEndpoint, expiredAt:" + seconds / 60 + "m left");
    } else {
        const seconds = expiredAt - currentTime;
        console.log("expiredAt:" + seconds / 60 + "m left");
    }

    if (!endpoint) {
        throw new Error("Endpoint not initialized");
    }

    const url = `https://${endpoint.r}.tts.speech.microsoft.com/cognitiveservices/v1`;
    const headers = {
        "Authorization": endpoint.t,
        "Content-Type": "application/ssml+xml",
        "User-Agent": "okhttp/4.5.0",
        "X-Microsoft-OutputFormat": outputFormat
    };
    const ssml = getSsml(text, voiceName, rate, pitch);
    const response = await fetch(url, {
        method: "POST",
        headers: headers,
        body: ssml
    });
    if (response.ok) {
        if (!download) {
            return response;
        }
        const resp = new Response(response.body, response);
        resp.headers.set("Content-Disposition", `attachment; filename="${crypto.randomUUID().replace(/-/g, "")}.mp3"`);
        return resp;
    } else {
        return new Response(response.statusText, { status: response.status });
    }
}

function getSsml(text: string, voiceName: string, rate: number, pitch: number): string {
    return `<speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" version="1.0" xml:lang="zh-CN"> <voice name="${voiceName}"> <mstts:express-as style="general" styledegree="1.0" role="default"> <prosody rate="${rate}%" pitch="${pitch}%" volume="50">${text}</prosody> </mstts:express-as> </voice> </speak>`;
}

async function voiceList(): Promise<any> {
    const headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.26",
        "X-Ms-Useragent": "SpeechStudio/2021.05.001",
        "Content-Type": "application/json",
        "Origin": "https://azure.microsoft.com",
        "Referer": "https://azure.microsoft.com"
    };
    const response = await fetch("https://eastus.api.speech.microsoft.com/cognitiveservices/voices/list", {
        headers: headers
    });
    return response.json();
}

async function hmacSha256(key: Uint8Array, data: string): Promise<Uint8Array> {
    const cryptoKey = await crypto.subtle.importKey(
        "raw",
        key,
        { name: "HMAC", hash: "SHA-256" },
        false,
        ["sign"]
    );
    const signature = await crypto.subtle.sign("HMAC", cryptoKey, encoder.encode(data));
    return new Uint8Array(signature);
}

async function base64ToBytes(base64: string): Promise<Uint8Array> {
    const binaryString = atob(base64);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    return bytes;
}

async function bytesToBase64(bytes: Uint8Array): Promise<string> {
    const base64 = btoa(String.fromCharCode.apply(null, bytes));
    return base64;
}

serve(handleRequest, { port: 8000 });

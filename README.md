# docker 部署

```shell
docker run -d -p 8080:8080 --name=tts zuoban/zb-tts
```


# cloudflare worker 部署
[worker.js](https://raw.githubusercontent.com/zuoban/tts/main/templates/worker.js)

支持接口
语音合成
/tts | GET / POST(json) try
参数列表：
1. t: 文本内容 (必填)
2. v: 语音名称 (可选), 默认为 zh-CN-XiaoxiaoMultilingualNeural
3. r: 语速 (可选), 默认为 0
4. p: 语调 (可选), 默认为 0
5. o: 输出格式 (可选), 默认为audio-24khz-48kbitrate-mono-mp3
声音列表
/voices | GET try
参数列表：
1. l: 语言区域 (可选), 使用 contains 匹配,如 l=zh
2. d: 显示详细信息 (可选) , 默认为 false, 如需显示详细信息, 请添加参数d , 如 /voices?d
package main

import "github.com/crosstyan/dumb_downloader/cmd"

// https://sxyz.blog/bypass-cloudflare-shield/
// https://developers.cloudflare.com/bots/concepts/ja3-fingerprint/
// https://www.zenrows.com/blog/bypass-cloudflare#active-bot-detection
// https://segmentfault.com/a/1190000041699815/en
// https://github.com/yolossn/JA3-Fingerprint-Introduction
// https://blog.csdn.net/chenzhuyu/article/details/132217262
func main() {
	cmd.Execute()
}
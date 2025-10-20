```
{timestamp(ID)}--{人間が見るコメント}__{任意の数のtag 1}_{任意の数のtag 1}.{拡張子}
20250903T083109--TCPIP入門__network_infra.pdf
```

## サポート

CLIツールとして、サポートする。

ディレクトリ一括。

- リネーム(ID付与)
  - go run . generate .
- バリデーション
  - go run . validate .

ファイル個別。

- タグカスタム
  - go run . tag ./20250903T083109--TCPIP入門__network_infra.pdf
  - タグ選択CLIを出す

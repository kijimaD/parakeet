```
{timestamp(ID)}--{人間が見るコメント}__{任意の数のtag 1}_{任意の数のtag 1}.{拡張子}
20250903T083109--TCPIP入門__network_infra.pdf
```

## サポート

CLIツールとして、サポートする。

ディレクトリ一括。

```
# ID付与
go run . generate . --ext pdf

# バリデーション
go run . validate . --ext pdf

# タグ編集
go run . tag {file名}
```

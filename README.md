### Запуск
```
$ git clone https://github.com/ArkovKonstantin/word-searcher.git
$ cd ./word-searcher
$ echo -e 'https://golang.org\n/etc/passwd\nhttps://golang.org\nhttps://golang.org' | go run main.go
   Count for /etc/passwd: 0
   Count for https://golang.org: 20
   Count for https://golang.org: 20
   Count for https://golang.org: 20
   Total: 60
```
### Запуск тестов
```
$ go test -v
```
![Image alt](https://github.com/ArkovKonstantin/word-searcher/raw/master/image.png)

{username} — ваш ник на ГитХабе;
{repository} — репозиторий где хранятся картинки;
{branch} — ветка репозитория;
{path} — путь к месту нахождения картинки.
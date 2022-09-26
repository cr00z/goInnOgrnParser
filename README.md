# INN/OGRN parser

First steps in Go

Realised:

* CSV input/output
* 'Worker Pool' pattern with cancellation and correct shutdown
* Loop queue for url paths (multithread safe)
* Web Crawler: net/http + [jackdanger/collectlinks](https://github.com/jackdanger/collectlinks])
* INN / OGRN simple checks

## Usage

First run:

`go build && ./goInnOgrnParser -i ../data/part.csv -o ../data/out.csv`

Help:

```
Usage of ./goInnOgrnParser:
    -a    get all numbers similar to inn and ogrn (lots of result = lots of garbage)
    -d    maximum parsing depth [default 1]
    -f    do not continue processing the link when the first result is received
    -h    http timeout, maximum server response time [default 10]
    -i    input csv file [default data/part.csv]
    -o    output CSV file, added result columns ('main_inn', 'main_ogrn', 'add_inn', 'add_ogrn') [default data/out.csv]
    -s    silence (ignored by verbose)
    -t    script execution limit in seconds [default 43200]
    -v    verbose output, added messages about processed urls, results, queue size, etc
    -w    worker pool size (number of threads) [default 10]
Use ^C for shutdown and save results.
Examples:
./goInnOgrnParser -a -d 3 -i in.csv -o out.csv -t 3600 -v -w 1000
```

## Example of work

![work](https://github.com/cr00z/goInnOgrnParser/blob/main/images/image1.png)

## Output

В исходную таблицу добавляются 5 столбцов:

* main_inn - главный ИНН (из результатов первого сайта)
* main_ogrn - главный ОГРН (из результатов первого сайта)
* add_inn - дополнительные ИНН (уникальный список через разделитель "," полученные с остальных сайтов)
* add_ogrn - дополнительные ОГРН (уникальный список через разделитель "," полученные с остальных сайтов)
* http_error - если при парсинге возникла ошибка, в это поле будет записано ее обозначение:
    * Response 4xx-5xx
	* No such host
	* Timeout
	* Connection refused
    * Connection reset by peer
    * Network is unreachable
    * Empty response
    * SSL Protocol error
    * Redirects
    * No route to host
    * для остальных ошибок - вся строка ошибки

## Some results

1. Все похожее на ИНН/ОГРН (много мусора), парсинг только стартовой страницы

`go build && ./goInnOgrnParser -a -d 1 -i ../data/part.csv -o ../data/test_d1a.csv -t 43200 -v -w 250`

```
INFO	2022/09/07 12:11:54 Finished work 117698, left 0 urls [ 0 paths ]
INFO	2022/09/07 12:12:03 Founded 18526 results (18 %)
INFO	2022/09/07 12:12:03 Took ================> 40m26.305081978s
```

2. Парсинг только стартовой страницы, на которой есть "ИНН" или "ОГРН"

`go build && ./goInnOgrnParser -d 1 -f -i ../data/part.csv -o ../data/test_d1.csv -t 43200 -v -w 250`

```
INFO	2022/09/07 12:55:54 Founded 5274 results (5 %)
INFO	2022/09/07 12:55:54 Took ================> 38m46.764405031s
```

3. Все похожее на ИНН/ОГРН (много мусора), ссылки до третьего уровня

`go build && ./goInnOgrnParser -a -d 3 -f -h 30 -i ../data/part.csv -o ../data/test_d3a.csv -t 43200 -v -w 1000`

```
INFO  2022/09/07 14:15:52 Finished work 987893, left 16071 urls [ 3736288 paths ]
INFO  2022/09/07 14:15:52 Pool received cancellation signal, closing result channel
INFO  2022/09/07 14:15:52 All workers done, shutting done!
INFO  2022/09/07 14:15:55 Founded 37690 results (37 %)
INFO  2022/09/07 14:15:55 Took ================> 12h1m20.795452977s
```

## Notes

Для запуска большого количества горутин может потребоваться увеличение лимита открытых файловых дескрипторов:

`ulimit -n 1048575`

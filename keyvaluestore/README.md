# Сравнение channels и mutexes
Написал простенькие шардированые Key-value хранилища. Одно на channels, другое на mutexes.

## Как запустить с профайлингом
Использовал следующие команды для бенчмаркинга:
```
go test -run BenchmarkKVChannel -bench=BenchmarkKVChannel -count 1 -cpuprofile BenchmarkKVChannel.cpu.out
go test -run BenchmarkKVMutexes -bench=BenchmarkKVMutexes -count 1 -cpuprofile BenchmarkKVMutexes.cpu.out
```

Эти команды получить SVG файл с деревом вызовов:
```
go tool pprof -svg ./BenchmarkKVMutexes.cpu.out > BenchmarkKVMutexes.cpu.svg
go tool pprof -svg ./BenchmarkKVChannel.cpu.out > BenchmarkKVChannel.cpu.svg
```

## Результаты
```
goos:  windows
goarch: amd64
cpu: Intel(R) Core(TM) i5-4460  CPU @ 3.20GHz
BenchmarkKVChannel-4     1631954               743.8 ns/op
BenchmarkKVMutexes-4     6860815               173.2 ns/op
```
Calls-grap для channel версии:
<a href="../docs/reports/BenchmarkKVChannel.cpu.svg"><img src="../docs/reports/BenchmarkKVChannel.cpu.svg" width="100%"></a>

Calls-grap для mutex версии:
<a href="../docs/reports/BenchmarkKVMutexes.cpu.svg"><img src="../docs/reports/BenchmarkKVMutexes.cpu.svg" width="100%"></a>

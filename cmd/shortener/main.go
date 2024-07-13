package main

import (
	"context"
	"fmt"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/server"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()
	s := server.New(cfg)

	// Создаём бесконечный контекст
	ctx, cancel := context.WithCancel(context.Background())
	// ожидаем завершение в горутине, отправляем в канал
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	// создаём WaitGroup
	g, gCtx := errgroup.WithContext(ctx)

	// запускаем сервер в горутине
	g.Go(func() error {
		return s.ListenAndServe()
	})
	// ждём завершения группы в горутине, выключаем сервер
	g.Go(func() error {
		<-gCtx.Done()
		// .Shutdown сначала перестаёт принимать новые запросы, обрабатывает текущие и выключается
		return s.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
}

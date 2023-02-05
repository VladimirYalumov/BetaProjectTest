Инструкция:  
сначала запустить docker-compose up -d  
go mod vendor  
log_consumer/main.go  
Далее voting/main.go и gamma/main.go (это просто заглушка)  
и далее запускаем основной main  

Все логи хранятся в монге, посчитал это хорошим ршение
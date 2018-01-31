# xy-inc

- Solução usa um banco de dados postgres, usando as tabelas e suas colunas para gerar a API.

- Para executar o projeto:

* O arquivo config.json deve ser configurado com o acesso a um banco postgres, no qual deve ter pelo menos uma tabela com algumas colunas.
* Já tem o executável compilado.
* Para executar deve acessar a pasta que esta o executável pelo cmd executar o mesmo passando como parâmetro o schema
 do banco que deseja gerar a API. <br>
Ex: > RESTgenerator.exe public

* Para compilar o projeto além do ambiente configurado para compilar o golang, baixe as bibliotecas:
> go get github.com/lib/pq <br>
> go get github.com/labstack/echo <br>
> go get github.com/labstack/echo/middleware <br> 

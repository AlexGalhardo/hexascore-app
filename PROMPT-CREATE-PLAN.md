# Contexto

Sou um desenvolvedor full stack com 10+ de experiência, apaixonado por games, livros, filmes e séries (tv shows).

Quero criar uma aplicação full stack, de longo prazo, privado, muito inspirado nos conceitos de 'database' + 'rede social simplificado' como fazem os sites: - <https://www.igdb.com/> - <https://www.imdb.com/> - <https://www.goodreads.com/>

Onde o foco vai ser o acesso da API dessas informações, com um developer experience world class, para que esse site se torna mais um referência na internet para ver (frontend) ou consumir esses dados (API)

O diferencial desse site é que os usuários não farão análises em textos, mas apenas em categorias usando notas de 0 a 100, parecido com o metacritic:

Exemplo: Jogo Wolverine 2026 PlayStation -> 5 categorias para dar notas em jogos: - Gameplay - Diversão - Qualidade Sonora - Qualidade Técnica - Qualidade Visual - Storytelling

E a nota final será dado com a média dessas 6 categorias. Assim o site fará um 'metascore' parecido com o metacritic com a base dos usuários, para avaliar bem os livros, games, filmes e séries.

## Detalhes

- O site terá 4 grandes seções de informação:
  - Livros
  - Jogos
  - Filmes
  - Séries (TV Shows)
- Os usuários

## Monetização

- Nos livros e jogos, em alguns deles, terão links de afiliados (principalmente amazon e mercadolivre) para os usuários poderem comprar os produtos, como forma de monetização.
- A API vai ser paga, com um plano free com limite de uso total (exemplo: conta free pode fazer até 100 requisição na API, depois precisa pagar)

## WEBSCRAPPING COM PYTHON

As principais fontes de dados que vamos pegar os dados para o banco de dados das informações serão:

- <https://howlongtobeat.com/>
- <https://www.igdb.com/>
- <https://imdb.com/>
- <https://www.themoviedb.org/>
- <https://www.metacritic.com/>
- <https://www.rottentomatoes.com/>
- <https://store.steampowered.com/>
- <https://www.goodreads.com/>
- <https://www.skoob.com.br/>
- <https://store.epicgames.com/>
- <https://www.gog.com/>

Os webscrapping vão ficar em outra pasta (criar uma pasta diferente chamada webscrappings, que vão ficar os códigos python separados do nextjs, da aplicação aqui que vai para o build da vercel)

Defina um padrão de como esses dados serão pegos de cada fonte de dado (seja scrapping ou via api aberta pública e grátis de cada site), como eles serão estruturados e salvos no banco de dados principal para serem renderizados no frontend

## Tech Stack principal

Vou usar fortemente a vercel para o deploy da aplicação para tudo, então:

- NextJS v16
- PostgresSQL
- DrizzleORM
- Tailwindcss v4
- <https://base-ui.com/>
- biomejs v2 para formatação
- ci/cd com github actions
- stripe para pagamentos com cartão de crédito (Brasil BRL R$, Euro, e USD)
- <https://www.abacatepay.com/llms.txt> para pagamentos com PIX
- resend e react-email para todos os emails da aplicação

Se houver mais outra ferramenta no ecosistema da vercel que você considera útil nesse projeto que eu nao citei, me diga para implementarmos nesse projeto.

Esse projeto será um grande monorepo, tudo centralizado em apenas um repositório.

## Detalhes

- i8n: EN (padrão) / PT-BR / ES espanhol
- Ter autenticação (use <https://better-auth.com/> ) com pagina de:
  - /login
    - Input password: minimo 8 characteres, máximo 32 characters -> ter pelo menos 1 character especial, 1 maiusculo, 1 minusculo e 1 numero
    - Input password ter o olhinho para ver a senha
    - Ter link no footer do formulário: 'Não tem conta ainda? Criar conta'
    - pode logar com conta social google também
  - /signup
    - Nickname da conta: minimo 4 characteres e máximo 16 characteres -> placeholder da primeira letra do nome enquanto digita ser em UPPERCASE
    - Input password: minimo 8 characteres, máximo 32 characters -> ter pelo menos 1 character especial, 1 maiusculo, 1 minusculo e 1 numero
    - Input password ter o olhinho para ver a senha
    - Ter link no footer do formulário: 'Já tem conta? Entrar'
    - pode criar conta social google também
      - crie uma senha forte automaticamente para quem criar conta com google, para o usuário poder mudar sua senha em /profile depois
  - /recover-password
  - /contact
    - Formulário de contato com 4 inputs: Nome e email (se usuário já logado, preenchidos, com inputs disabled mas que podem ser enviados no post da api) -> select com 'Dúvidas', 'Sugestões & Feedback', 'Bugs e Problemas Técnicos' e input de textarea com 5 rows (minimo 32 characteres, máximo 512 com essa informação mostrando no frontend 0/512 characteres)
  - /reset-password
  - / landing page:
    - Ter um navbar principal que pode buscar qualquer item (livro, série, filme, jogo) e com um catalogo das 4 seções abaixo, para os usuários que entrarem no site já saberem do que se trata;
  - /policy -> pagina de política de privacidade
  - /terms -> página de termos de uso e cookies
  - /profile -> página com dados do usuário: nome, email, mudar senha, 2fa, deletar conta (30 dias para deletar -> usar modal com coldown de 10 segundos informando isso para confirmar deleção)
  - /profile/api/checkout -> assinar algum plano pago para consumir a API (a definir no planejamento e brainstorm)

### Tech Stack desenvolvimento local

- Usaremos também SQLITE
- Docker & docker-compose

## Frontend

- Deve ser parecida com o IMDB, com bastante informação, informativo, didático, poucas imagens, 2-3 trailers oficiais no máximo, nota final média dos usuários e de outras plataformas e modal mostrando notas de outros usuários do site;
- Obrigatoriamente todas as telas desse projeto devem ser mobile friendly;
- Ter toggle na navbar com light/dark theme;
- Coloque o banner de aceitar cookies e salve o estado do usuário no browser (theme, cookies, outras infos úteis)

## API REST

Na conta do /profile/api vai ter uma documentação completa usando <https://scalar.com/products/api-references/integrations/react> para o usuário ver os endpoints da api para ele consultar as informações -> planos free no máximo 10 itens por consulta de uma vez - buscar por id - buscar por nome - buscar por categoria - buscar por nota final ou especificas - etc

A API deve ser bem completa, extremamente DX DEVELOPER FRIENDLY EXPERIENCE (Pegue como referência a excelente developer experience que a <https://resend.com/docs/introduction> tem com os desenvolvedores)

A API deve seguir as melhores práticas atuais, e com endpoints começando com v1 na primeira versão que estamos criando - <https://www.openapis.org/> - <https://swagger.io/specification/>

## Testes, Seeds, Q.A, Pentest

- Para esse projeto ter uma base boa para ir para produção na versão 1.0.0 estável, deve ter pelo menos uns 100.000 itens já cadastrados e terá um seed de testes com 10.000 usuários cadastrados com notas nos itens cadastrados;
- Também crie scripts e um fluxo para ter um 'Q.A/Pentest' simulando um humano real profissional que vai testar todas as telas, urls, endpoints, botões, fluxos, edge cases, fallbacks, etc

## Boas práticas e engenharia de software

- Se atente no planejamento e desenvolvimento desse projeto em prevenir todos os: <https://owasp.org/projects/top-ten>
- Crie branch dev que vai simular ambiente 'sandbox' onde todos os checks do CI/CD tem que passar para seguir com push na main
- Use as boas práticas como:
  - semver
  - conventional commits
  - keepachangelog
  - releases no github começando na versão 0.0.1 e ir evoluindo a cada fix/hotfix/patch etc
- Deixe explicito no CLAUDE/AGENTS.md e nos rules do .claude/rules/ -> sempre usar os padrões do andrej-karpathy (https://github.com/multica-ai/andrej-karpathy-skills/blob/main/skills/karpathy-guidelines/SKILL.md) ponytail e graphify para otimização e evolução desse código durante o desenvolvimento;
- Sempre use apenas as ultimas versões ESTÁVEIS prontas para produção dos pacotes no package.json (nada de release, beta, latest, etc) -> deixe as versões fixadas;

## ANTES DE COMEÇAR

- Você percebeu que tem alguns arquivos e pastas aqui nesse projeto já. Eu peguei de outro projeto que eu estava refatorando, também com nextjs, com algumas estruturas que eu usei lá. Quero que você veja tudo para ter uma noção/referência do que você vai fazer aqui

1. Antes de qualquer coisa, use a skill: 'superpower-brainstorming' para fazer um brainstorming final para finalizar detalhes dessa implementação (nome final da aplicação, dns, padrão de rotas, detalhes de frontend, etc);
2. Depois use a skill grilling ou interview-me e me faça todas as perguntas necessárias: dúvidas, implementação, sugestões, dicas, tech stack, etc para criarmos o arquivo principal desse projeto: PLAN.md que vai ficar todo o planejamento detalhado dessa aplicação para ser implementado.

## System design e documentações

- Quando necessário e útil, use a lib https://reactflow.dev/ para melhorar documentações.

## Worktrees

- Depois de feito o PLAN.md e começando a implementação, quando necessário, crie worktrees para usar agentes em trabalhos paralelos que não conflitem com o trabalho da sessão princial do chat (criar webscrappings do python, documentações, etc)

## GIT

- Comece um git do zero e use o remote: git remote add origin <https://github.com/AlexGalhardo/claude-code-internet-api.git>
- Automatize ao máximo todos os fluxos desse projeto: git, ci/cd github actions, stripe, docker, abacatepay, vercel, etc -> use CLI e MCPs aqui já configurados para você ter o máximo de autonomia possível para trabalhar

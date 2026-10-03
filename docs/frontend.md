# Frontend e rotas


Locale: EN sem prefixo, `/pt-br/…`, `/es/…`. Slugs de rota em inglês nos 3 idiomas. Tudo mobile-first, light/dark.

| Rota | Conteúdo |
| --- | --- |
| `/` | navbar com busca global (autocomplete) + blocos por seção (melhores, lançamentos, mais avaliados) |
| `/games` `/movies` `/series` `/books` | catálogo denso com filtros na URL |
| `/{type}/[slug]` | página do item |
| `/search?q=` | resultados entre tipos |
| `/u/[nickname]`, `/u/[nickname]/lists/[id]` | perfil público, listas |
| `/feed` | avaliações de quem você segue |
| `/pricing` | planos da API |
| `/docs/*` | guias + Scalar |
| `/login` `/signup` `/recover-password` `/reset-password` `/verify-email` | auth |
| `/profile` | dados, senha (definir/mudar), 2FA, passkeys, excluir conta |
| `/profile/api` · `/profile/api/checkout` | chaves, uso, Scalar · assinar plano/comprar créditos |
| `/contact` `/policy` `/terms` | contato, privacidade, termos e cookies |
| `/admin/*` | contato, denúncias, afiliados, edição/merge de itens, status da ingestão, banir usuário (só EN) |

**Página do item**: cabeçalho (título, ano, tipo, capa pequena, gêneros/plataformas) · metascore grande + radar hexagonal das 6 categorias + nº de avaliações · botão "ver notas dos usuários" (modal paginado, só perfis públicos) · formulário com 6 controles (slider + campo numérico, acessível por teclado, média ao vivo) · notas externas (MC, RT, IMDb, HLTB) como links com atribuição · até 3 trailers (`youtube-nocookie`, carregam só no clique) · tabela de metadados · status + adicionar à lista · "onde comprar" (livros/jogos).

**Formulários** (validação zod idêntica cliente/servidor):
- Senha: 8–32, ≥ 1 maiúscula, 1 minúscula, 1 número, 1 especial; botão olho. Rodapé do login: "Não tem conta ainda? Criar conta"; do signup: "Já tem conta? Entrar". Ambos com Google.
- Nickname: 4–16 `[a-zA-Z0-9_]`, único; avatar placeholder mostra a 1ª letra em MAIÚSCULA enquanto digita.
- Contato: nome + e-mail preenchidos e `disabled` se logado (servidor usa a sessão); select "Dúvidas" / "Sugestões & Feedback" / "Bugs e Problemas Técnicos"; textarea 5 linhas, 32–512 caracteres com contador `0/512`.
- Excluir conta: modal explicando os 30 dias, botão confirmar desabilitado por 10s com contagem.

**Preferências**: tema, consentimento e locale em cookies próprios; tema lido no servidor (sem flash, sem lib). Banner de cookies: só essenciais (Analytics é cookieless) + consentimento para embeds do YouTube.

**UI**: Base UI + Tailwind 4 (tokens por seção), `next-intl`, radar em SVG puro.

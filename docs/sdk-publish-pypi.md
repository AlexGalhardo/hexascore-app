# Publicar o SDK Python no PyPI

Passo a passo para publicar `hexascore` (pasta `sdks/python`) em <https://pypi.org>, o registro de pacotes do ecossistema Python (é de lá que `pip install`, `uv add` e `poetry add` baixam). O SDK TypeScript está em [`sdk-publish-npm.md`](sdk-publish-npm.md) e [`sdk-publish-jsr.md`](sdk-publish-jsr.md).

| | |
| --- | --- |
| Pacote | `hexascore` |
| Pasta | `sdks/python` |
| O que é publicado | `dist/hexascore-<versão>.tar.gz` (sdist) e `dist/hexascore-<versão>-py3-none-any.whl` (wheel) |
| Versão | campo `version` de `sdks/python/pyproject.toml` (hoje `0.1.0`) |
| Ferramenta | `uv` (build e publicação) |
| Estado | ainda não publicado; o nome `hexascore` estava livre em 2026-10-10 |

## Como a pasta é organizada

```text
sdks/python/
├─ pyproject.toml   escrito à mão: nome, versão, metadados, dependências
├─ README.md        vira a descrição da página no PyPI
├─ LICENSE
├─ config.yml       configuração do gerador
├─ tests/           testes de fumaça (transporte simulado, sem rede)
└─ hexascore/       pacote GERADO (bun run sdk:generate:python): nunca editar
```

O gerador roda com `--meta none`, então só reescreve `hexascore/`. Nome e versão do pacote ficam no `pyproject.toml` e não se perdem ao regenerar.

## 1. Preparar as contas (uma vez)

O PyPI tem dois registros separados, com contas separadas:

| Registro | Para quê | Cadastro |
| --- | --- | --- |
| TestPyPI | ensaiar a publicação sem consequência | <https://test.pypi.org/account/register/> |
| PyPI | o registro de verdade | <https://pypi.org/account/register/> |

Em cada um:

1. Crie a conta e confirme o e-mail.
2. Ative 2FA (*Account settings → Two factor authentication*). É obrigatório para publicar.
3. Gere um token de API: *Account settings → API tokens → Add API token*.
   - Na **primeira** publicação o projeto ainda não existe, então o escopo tem de ser *Entire account*.
   - Depois da primeira publicação, apague esse token e crie outro com escopo só do projeto `hexascore`.
   - O token começa com `pypi-` e aparece uma única vez.

Guarde os tokens fora do repositório. Para a sessão atual do terminal:

```sh
export UV_PUBLISH_TOKEN="pypi-..."          # bash / Git Bash
$env:UV_PUBLISH_TOKEN = "pypi-..."          # PowerShell
```

Com token, o usuário é sempre `__token__`; o `uv publish` cuida disso.

## 2. Conferir o pacote antes de publicar

Na raiz do repositório:

```sh
bun run openapi:check          # o contrato commitado está em dia com o código
bun run sdk:generate:python    # regenera sdks/python/hexascore (precisa do uv)
```

Em `sdks/python`:

```sh
uv run pytest -q               # 2 testes
rm -rf dist                    # PowerShell: Remove-Item -Recurse -Force dist
uv build                       # gera dist/*.tar.gz e dist/*.whl
uvx twine@6.2.0 check dist/*   # valida os metadados e o README
```

Teste o wheel num ambiente limpo, sem o código-fonte por perto:

```sh
uv run --no-project --isolated --with ./dist/hexascore-0.1.0-py3-none-any.whl \
  python -c "import hexascore; from hexascore.api.games import list_games; print('ok')"
```

## 3. Escolher a versão

Uma versão publicada no PyPI **nunca pode ser reenviada**, nem depois de apagada. Errou? Só subindo o número.

Edite `version` em `sdks/python/pyproject.toml` (ou `uv version 0.1.1`). Use a mesma versão de `sdks/ts/package.json` e `sdks/ts/jsr.json`. Regra de SemVer em [`sdk-publish-npm.md`](sdk-publish-npm.md#3-escolher-a-versão).

Pré-release em Python: `0.2.0b1` (beta), `0.2.0rc1` (candidata). O `pip` só instala pré-release com `--pre`.

## 4. Ensaiar no TestPyPI

```sh
cd sdks/python
uv publish --publish-url https://test.pypi.org/legacy/ --token "$TEST_PYPI_TOKEN"
```

Confira a página <https://test.pypi.org/project/hexascore/> (README renderizado, versão, links) e instale a partir de lá. As dependências (`httpx`, `attrs`) vêm do PyPI real:

```sh
uv run --no-project --isolated \
  --index https://test.pypi.org/simple/ --index-strategy unsafe-best-match \
  --with hexascore==0.1.0 python -c "import hexascore; print('ok')"
```

## 5. Publicar no PyPI

```sh
cd sdks/python
uv publish          # lê UV_PUBLISH_TOKEN; envia tudo que está em dist/
```

Conferir:

```sh
uv run --no-project --isolated --with hexascore==0.1.0 python -c "import hexascore; print('ok')"
```

e a página <https://pypi.org/project/hexascore/>.

## 6. Depois de publicar

1. Troque o token de conta inteira por um token só do projeto `hexascore` (passo 1.3).
2. Marque o commit: `git tag sdk-v0.1.0 && git push origin sdk-v0.1.0` (a mesma tag do SDK TypeScript, se as versões forem iguais).
3. Registre no `CHANGELOG.md` e marque a pendência no `PLAN.md`.

## Publicar por automação (opcional)

Com o GitHub Actions desativado neste repositório, a publicação é manual. Quando for automatizar, use **Trusted Publishing** em vez de token:

1. No PyPI: *Your projects → hexascore → Publishing → Add a new publisher → GitHub*.
2. Informe dono `AlexGalhardo`, repositório `hexascore-app`, o nome do arquivo do workflow e o environment (ex.: `pypi`).
3. O job precisa de `permissions: id-token: write` e roda `uv publish` sem token nenhum.

Dá para cadastrar o publisher antes do projeto existir (*pending publisher*), o que dispensa o token de conta inteira até na primeira publicação.

## Erros comuns

| Erro | Causa | Solução |
| --- | --- | --- |
| `400 File already exists` | essa versão já foi enviada | suba a versão e gere o `dist/` de novo |
| `403 Invalid or non-existent authentication` | token errado, ou token do TestPyPI usado no PyPI | confira o token e o registro |
| `403 ... isn't allowed to upload to project` | token com escopo de outro projeto | gere token do projeto certo |
| `400 The name 'hexascore' is too similar to an existing project` | nome reservado ou parecido demais | escolha outro nome em `pyproject.toml` |
| Envia versões antigas junto | `dist/` com arquivos de builds anteriores | apague `dist/` antes do `uv build` |
| README sem formatação na página | Markdown inválido para o PyPI | `uvx twine@6.2.0 check dist/*` mostra o problema |

## Corrigir uma publicação errada

- **Yank (recomendado):** na página do projeto, *Manage → Releases → Options → Yank*. A versão continua instalável por quem a fixou (`==0.1.1`), mas deixa de ser escolhida por padrão.
- **Apagar:** possível, mas o número da versão fica bloqueado para sempre e quebra quem dependia dela. Prefira yank e publique a correção.

## Checklist

- [ ] `bun run openapi:check` e `bun run sdk:generate:python` sem diferenças não commitadas
- [ ] versão igual em `pyproject.toml`, `package.json` e `jsr.json`
- [ ] `uv run pytest -q` passando
- [ ] `dist/` apagado, `uv build`, `twine check`
- [ ] wheel instalado e importado em ambiente limpo
- [ ] ensaio no TestPyPI
- [ ] `uv publish` e instalação a partir do PyPI
- [ ] token de conta trocado por token do projeto
- [ ] tag `sdk-vX.Y.Z`, `CHANGELOG.md`, `PLAN.md`

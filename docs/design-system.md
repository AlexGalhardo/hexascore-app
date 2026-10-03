# Design system

Direção: **dados editorial** — um almanaque de referência, denso e informativo, poucas imagens. Implementação: Tailwind 4 (tokens em CSS) + Base UI (primitivos acessíveis sem estilo).

## Princípios

- Informação antes de decoração: tabelas e listas densas, hierarquia por tipografia, não por cartões e sombras.
- Números são protagonistas: metascore e categorias em fonte tabular, sempre alinhados.
- Mobile-first: toda tela funciona em 360 px; tabelas viram listas empilhadas abaixo de `md`.
- Light e dark com o mesmo contraste mínimo (WCAG AA, 4.5:1 para texto).

## Tipografia

| Papel | Fonte | Uso |
| --- | --- | --- |
| Display | serif (definida no 0.0.x via `next/font`) | títulos de item e de página |
| Texto | sans | corpo, UI, formulários |
| Números | mono / `font-variant-numeric: tabular-nums` | notas, contagens, datas |

## Cor

- Neutros para superfície e texto (tokens `--bg`, `--fg`, `--muted`, `--border`), redefinidos em `[data-theme="dark"]`.
- **Acento por seção** (um único por seção, usado em links ativos, foco e marcadores): jogos, filmes, séries, livros — valores definidos no 0.0.x.
- **Escala de nota** (estilo Metacritic), aplicada a metascore e categorias:

| Faixa | Significado | Token |
| --- | --- | --- |
| 0–49 | negativa | `--score-low` (vermelho) |
| 50–74 | mista | `--score-mid` (amarelo) |
| 75–100 | positiva | `--score-high` (verde) |
| sem nota (< 5 avaliações) | insuficiente | `--score-none` (neutro) |

Cor nunca é o único sinal: o número sempre aparece junto.

## Componentes assinatura

- **Radar hexagonal**: SVG puro, 6 eixos (categorias da seção), 0–100, rótulos com valor; usado na página do item e no modal de notas.
- **Score badge**: número grande + cor da faixa; tamanhos `sm`/`md`/`lg`.
- **Score input**: slider + campo numérico sincronizados (0–100), operável por teclado (setas ±1, PageUp/PageDown ±10).
- **Data table**: tabela densa com cabeçalho fixo e versão empilhada em mobile.

## Acessibilidade

Foco visível em todos os controles, alvos de toque ≥ 44 px, `prefers-reduced-motion` respeitado, modais com foco preso (Base UI `Dialog`), `lang` correto por locale.

Este documento cresce junto com `apps/web/components/`; toda nova peça de UI reutilizável entra aqui.

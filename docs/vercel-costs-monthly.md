# Custos mensais de produção (estimativa)

Estimativa de quanto custa rodar o Hexascore em produção: Vercel Pro, Vercel Blob (capas), Neon e Upstash via Vercel Marketplace, Resend, domínio `hexascore.app`, GitHub Actions, Stripe e AbacatePay.

- **Todos os preços foram conferidos em 2026-10-03** nas páginas oficiais linkadas em cada seção (os docs da Vercel mostram `last_updated: 2026-09-14`). Valores em USD sem impostos (a Vercel cobra VAT/impostos à parte, conforme o endereço de cobrança).
- **Câmbio usado:** PTAX venda do Banco Central de 2026-10-02 (último dia útil antes de 2026-10-03), **US$ 1 = R$ 5,2238** ([API PTAX do BCB](https://olinda.bcb.gov.br/olinda/servico/PTAX/versao/v1/odata/)). A fatura do cartão em BRL ainda soma IOF e spread do banco, que ficaram de fora.
- Volumes de tráfego são **premissas**, não medições. Depois do lançamento, troque pelas métricas reais de **Usage** (Vercel), **Storage → Usage** (Neon) e do console da Upstash.

## Dá para fazer upgrade de Neon e Upstash estando no Vercel Pro?

**Sim.** As duas são *native integrations* do Vercel Marketplace e funcionam em qualquer plano da Vercel, inclusive o Pro ([Vercel: product integrations](https://vercel.com/docs/integrations/install-an-integration/product-integration)). Os fatos:

- **A cobrança passa pela Vercel.** No Neon "Vercel-Managed Integration", "all billing [is routed] through your Vercel invoice" e plano, tier e compute só se mudam no dashboard da Vercel ([Neon docs](https://neon.com/docs/guides/vercel-managed-integration)). Na Upstash criada pela Vercel ("Create New Upstash Account"), pagamento, plano e regiões também ficam no dashboard da Vercel ([Upstash docs](https://upstash.com/docs/redis/howto/vercelintegration)). O Marketplace lista para Upstash Redis os planos *Free, Pay as You Go, Fixed* ([vercel.com/marketplace/upstash](https://vercel.com/marketplace/upstash)), e para Neon "Plans starting at $0" ([vercel.com/marketplace/neon](https://vercel.com/marketplace/neon)).
- **É uma cobrança separada da assinatura Pro.** Cada integração tem o próprio ciclo de fatura e pode usar outro método de pagamento ([Vercel: integration billing](https://vercel.com/docs/integrations/create-integration/billing)). Os docs não dizem que o crédito mensal de US$ 20 do Pro cobre produtos do Marketplace, então esta estimativa **assume que não cobre**.
- **Passos no dashboard (Neon):** Vercel → time → **Storage** → banco Neon → **Settings → Change Configuration** → escolher o plano (Launch) e o tamanho de compute/scale-to-zero → confirmar. A troca de plano vale para **todos os bancos dessa instalação**. Outro caminho: **Integrations → Neon → Manage**.
- **Passos no dashboard (Upstash):** Vercel → **Storage** → banco Upstash → **Settings** → trocar o plano (Free → Pay as You Go ou Fixed) → confirmar. Em Pay as You Go, defina um **budget** mensal: ao passar dele o banco recebe rate limit em vez de gerar cobrança ([Upstash pricing FAQ](https://upstash.com/pricing/redis)). Nos planos Fixed, deixe o **auto-upgrade** desligado se quiser um teto rígido.
- **Quem faz o upgrade:** trocar de plano é uma **mudança de cobrança**. O Claude consegue preparar e explicar os passos (e até rodar `vercel integration add … --plan …` para um recurso *novo*), mas não confirma cobranças em seu nome. O upgrade precisa ser confirmado por você, no dashboard, como Owner/Billing do time.

## Como o app gasta recursos (do código)

| Fato do repositório | Onde | Efeito no custo |
| --- | --- | --- |
| Toda página `[locale]` é renderizada dinamicamente (nonce da CSP) | `CHANGELOG` 0.0.16 | cada page view invoca uma Function (sem ISR) e lê o Runtime Cache (`unstable_cache`) |
| Capas com `<img>` simples, sem `next/image` | `app/[locale]/[section]/[slug]/page.tsx` | **Image Optimization não é usado: US$ 0** |
| Capas copiadas para o Blob público em `covers/<item_id>.<ext>` pelo `covers.py` (GitHub Actions) | `docs/ingestion.md` | upload = 1 Advanced Operation por arquivo; feito de fora da Vercel, não gera Fast Data Transfer |
| 3 crons Vercel diários (`flush-usage`, `billing-reminders`, `purge-accounts`) | `apps/web/vercel.json` | dezenas de invocações por mês (desprezível) |
| Sem `regions` no `vercel.json` | `apps/web/vercel.json` | Functions na região padrão **iad1 (Washington)**; os preços abaixo usam iad1 |
| Cota da API em Redis: throttle `slidingWindow` + `INCR` + `HINCRBY`/`EXPIRE` por request | `apps/web/lib/api/quota.ts` | **~4–5 comandos Redis por request da API** (créditos/refund somam 2–4) |
| Planos da API: Free 100 req no total, Pro 50k/mês, Business 500k/mês | `PLAN.md` | define o teto de comandos Redis |
| Catálogo alvo de 100k itens; seed de escala com 10k usuários e 2.175.836 avaliações | `docs/ingestion.md`, `docs/testing.md` | dimensiona o Neon |
| Ingestão: diária (teto de 180 min), semanal (350 min), bootstrap manual (6 jobs de até 6 h) | `.github/workflows/ingest-*.yml` | minutos de GitHub Actions (o repositório é **privado**) |
| Vercel Web Analytics 2.0.1 + Speed Insights 2.0.0 | `CHANGELOG` 0.0.16 | Web Analytics é cobrado por evento no Pro, Speed Insights (free tier) não |

## 1. Vercel Pro

Fontes: [Pro plan](https://vercel.com/docs/plans/pro-plan), [Pricing](https://vercel.com/docs/pricing), [Flat Rate CDN](https://vercel.com/docs/pricing/flat-rate-cdn), [Fluid compute pricing](https://vercel.com/docs/functions/usage-and-pricing), [iad1 pricing](https://vercel.com/docs/pricing/regional-pricing/iad1), [gru1 pricing](https://vercel.com/docs/pricing/regional-pricing/gru1), [Web Analytics](https://vercel.com/docs/analytics/limits-and-pricing), [Speed Insights](https://vercel.com/docs/speed-insights/limits-and-pricing). Conferido em 2026-10-03.

**Modelo atual:** não existem mais cotas incluídas por recurso no Pro (como "1 TB de banda" ou "1.000 GB-h"). O Pro paga **US$ 20/mês**, que já inclui 1 assento e **US$ 20 de crédito de uso**. Esse crédito abate o uso on-demand e o que passar dele é cobrado. A CDN tem preço próprio, o **Flat Rate CDN**.

| Item | Pro (iad1) | gru1 (São Paulo), comparação |
| --- | --- | --- |
| Taxa da plataforma | US$ 20/mês (1 assento + US$ 20 de crédito de uso) | — |
| Assento extra | US$ 20/mês | — |
| Flat Rate CDN (CDN requests + Fast Data Transfer + **Blob Data Transfer**) | incluso: **1M requests + 1 TB/mês**; US$ 20/mês: 10M + 50 TB; US$ 100: 50M + 50 TB; US$ 300: 150M + 50 TB | — |
| CDN Requests on-demand | US$ 2,00 / 1M | US$ 3,20 / 1M |
| Fast Data Transfer on-demand | US$ 0,15 / GB | US$ 0,22 / GB |
| Fast Origin Transfer (sempre usage-based) | US$ 0,06 / GB | US$ 0,41 / GB |
| Fluid Active CPU | US$ 0,128 / hora | US$ 0,221 / hora |
| Fluid Provisioned Memory | US$ 0,0106 / GB-h | US$ 0,0183 / GB-h |
| Invocations | US$ 0,60 / 1M | — |
| ISR / Runtime Cache, leituras | US$ 0,40 / 1M read units | US$ 0,64 |
| ISR / Runtime Cache, escritas | US$ 4,00 / 1M write units | US$ 6,40 |
| Image Optimization (não usado) | US$ 0,05 / 1K transformações; US$ 0,40 / 1M cache reads; US$ 4,00 / 1M cache writes | US$ 0,0812 / 1K |
| Web Analytics | **nenhum evento incluso no Pro**; US$ 0,03 / 1K eventos (Plus: +US$ 10/mês) | — |
| Speed Insights | **grátis** em todos os planos (10.000 eventos / 30 dias por time, sem cobrança extra). Speed Insights Plus: US$ 10/projeto/mês + US$ 0,65 / 10K eventos | — |
| Builds | Elastic (padrão em times novos): US$ 0,0035 / CPU-minuto; Basic: US$ 0,007/min; Standard só é cobrado com Elastic ou on-demand concurrency | — |
| Observability Plus (opcional) | US$ 1,20 / 1M eventos | — |

Observações:

- **Flat Rate CDN:** quando a média mensal passa da capacidade do tier, a Vercel sobe o tier no ciclo seguinte (um pico de um dia não conta). A página mostra um toggle em **Settings → Billing → Flat Rate CDN**, então confira se ele está ligado no time. Desligado, CDN requests, FDT e Blob Data Transfer são cobrados on-demand (linhas acima). A política de *fair use* exclui "large-scale delivery of media… where such delivery constitutes a majority of bandwidth": se as capas virarem a maior parte da banda, a Vercel pode mover o projeto para a Flex CDN.
- Não ficou claro na documentação se o crédito de US$ 20 abate o preço de um tier pago do Flat Rate CDN. Aqui o tier entra **somado** ao valor.

## 2. Vercel Blob (capas)

Fonte: [Vercel Blob pricing](https://vercel.com/docs/vercel-blob/usage-and-pricing) + tabelas regionais acima (conferido em 2026-10-03). Preços iad1 (gru1 entre parênteses):

| Recurso | Preço |
| --- | --- |
| Armazenamento | US$ 0,023 / GB-mês (US$ 0,041) |
| Simple Operations (cache MISS de download, `head`) | US$ 0,40 / 1M (US$ 0,56). **Cache HIT não conta** |
| Advanced Operations (`put`, `list`, `copy`, multipart = início + 1 por parte + fim; navegar no dashboard também conta) | US$ 5,00 / 1M (US$ 7,00). `del()` é grátis |
| Blob Data Transfer (todo download) | US$ 0,05 / GB (US$ 0,073), ou **coberto pelo Flat Rate CDN** |
| Fast Origin Transfer (só em cache MISS) | US$ 0,06 / GB (US$ 0,41) |
| CDN Request por download | conta no Flat Rate CDN, ou US$ 2,00 / 1M on-demand |

Limites do Pro: 7.200 simple ops/min e **4.500 advanced ops/min (75/s)**. Um upload em massa de 1M capas leva no mínimo ~3,7 h, então o `covers.py` precisa respeitar 75 put/s.

### 2.1 Um milhão de capas em WebP otimizado

Armazenamento (1 GB = 10⁹ bytes; miniatura opcional de 5 KB por capa):

| Tamanho médio | GB (só capa) | GB (+ miniatura) | Armazenamento/mês iad1 | (+ miniatura) | gru1 (+ miniatura) |
| --- | --- | --- | --- | --- | --- |
| 30 KB | 30 | 35 | **US$ 0,69** | US$ 0,81 | US$ 1,44 |
| 60 KB | 60 | 65 | **US$ 1,38** | US$ 1,50 | US$ 2,67 |
| 100 KB | 100 | 105 | **US$ 2,30** | US$ 2,42 | US$ 4,31 |

Upload único (uma `put` por arquivo, sem multipart): **1M capas = US$ 5,00**; com miniaturas (2M puts) = **US$ 10,00** (gru1: US$ 7 / US$ 14). Reenviar a mesma capa em toda ingestão cobra de novo, então o `covers.py` deve pular o que já existe (comparar com o `coverUrl` gravado em vez de `head` ou `list`, que também são operações cobradas).

Entrega mensal, supondo **90% de cache HIT** na CDN (capas são imutáveis por `item_id`) e que cada view baixa a capa completa:

| Views/mês | Tamanho | Transferência | **On-demand** (CDN req + BDT + FOT + ops) | **Com Flat Rate CDN** (só FOT + ops; req e BDT entram no tier) |
| --- | --- | --- | --- | --- |
| 100 mil | 30 / 60 / 100 KB | 3 / 6 / 10 GB | US$ 0,37 / 0,54 / 0,76 | US$ 0,02 / 0,04 / 0,06 |
| 1 milhão | 30 / 60 / 100 KB | 30 / 60 / 100 GB | US$ 3,72 / 5,40 / 7,64 | US$ 0,22 / 0,40 / 0,64 (mas 1M requests já esgota o tier incluso) |
| 10 milhões | 30 / 60 / 100 KB | 0,3 / 0,6 / 1 TB | US$ 37,20 / 54,00 / 76,40 | US$ 2,20 / 4,00 / 6,40 **+ tier de US$ 20 (10M req) ou US$ 100 (50M)** |

Conclusão: **guardar 1M capas custa de US$ 0,69 a 2,42 por mês**, mais US$ 5–10 de upload uma única vez. O que pesa é a **entrega**, porque cada view é um CDN request. Com 60 KB e 1M views/mês o custo fica entre ~US$ 5 (on-demand) e ~US$ 0,40 (dentro de um tier Flat Rate). Usar a miniatura de 5 KB nas listagens corta a transferência das grades em ~90%.

## 3. Neon Postgres (Marketplace)

Fonte: [Neon plans](https://neon.com/docs/introduction/plans) e [Vercel-Managed Integration](https://neon.com/docs/guides/vercel-managed-integration) (conferido em 2026-10-03). A Neon chama o banco gerenciado pela Vercel de "Lakebase Postgres". Os planos e preços abaixo são os da Neon. A página da Vercel diz só "Plans starting at $0", então **confirme no seletor de plano** se os valores no Marketplace são os mesmos.

| | Free | **Launch** | Scale |
| --- | --- | --- | --- |
| Mensalidade | US$ 0 | sem mínimo, paga pelo uso | sem mínimo, paga pelo uso |
| Compute | 100 CU-h/projeto/mês | **US$ 0,106 / CU-h** | US$ 0,222 / CU-h |
| Autoscaling | até 2 CU (8 GB RAM) | até 16 CU | até 16 CU, ou fixo até 56 CU |
| Scale to zero | 5 min, não desliga | 5 min, pode desligar | 1 min até sempre ligado |
| Storage | **1 GB/projeto** (20 GB por conta) | US$ 0,35 / GB-mês, ilimitado | US$ 0,35 / GB-mês |
| Instant restore (histórico) | 6 h, até 1 GB-mês | US$ 0,20 / GB-mês, até 7 dias | US$ 0,20 / GB-mês, até 30 dias |
| Egress | 5 GB | 500 GB, depois US$ 0,10/GB | 500 GB, depois US$ 0,10/GB |
| Branches | 10/projeto | 10 + US$ 1,50/branch-mês | 25 + US$ 1,50/branch-mês |
| Extra | — | — | SLA, compliance, private networking |

1 CU ≈ 4 GB de RAM. Faturas abaixo de US$ 0,50 não são cobradas.

### 3.1 Tamanho do banco (100k itens, 10k usuários, 2,2M avaliações)

Estimativa por tamanho de linha (cabeçalho do tuple de 24 B, páginas de 8 KB, fillfactor 100) a partir de `apps/web/lib/db/schema/*`:

| Tabela | Conta | Estimativa |
| --- | --- | --- |
| `ratings` (2,2M linhas: id de usuário em texto, uuid, 7 smallint, 2 timestamptz ≈ 115 B) | heap ~250 MB + PK (user, item) ~155 MB + `ratings_item_idx` ~80 MB + `ratings_user_idx` ~135 MB | **~0,6 GB** |
| `items` (100k, `meta`/`external_ids` jsonb, 4 btree + 2 GIN) | ~150 MB + índices ~100 MB | ~0,25 GB |
| `item_translations` (300k = 3 idiomas; sinopse + tsvector + GIN tsvector + trigram) | ~480 MB + ~180 MB | ~0,65 GB |
| auth/social/billing (10k usuários, sessões, status, follows, ledger) | — | < 0,1 GB |
| **Subtotal sem payload bruto** | | **~1,5–1,8 GB** |
| `source_records.payload` (JSON bruto do Wikidata/Wikipedia/Open Library/Steam/GOG, TOAST comprimido) | não medido; entidades Wikidata têm dezenas de KB | **+1,5 a 4 GB** |

**Total estimado: 3–6 GB.** Isso já não cabe no Free (1 GB por projeto). Para medir de verdade: `select pg_size_pretty(pg_database_size(current_database()));` e `pg_total_relation_size('source_records')` depois do bootstrap. Se o payload bruto for o maior item, compense podá-lo depois de normalizar (o snapshot já fica no GitHub Release).

### 3.2 Recomendação

- **Launch** em produção. O Free não comporta o catálogo e não permite desligar o scale-to-zero. O Scale (o dobro do preço por CU-h) só vale quando houver necessidade de SLA, PITR de 30 dias ou compliance.
- Lançamento: autoscaling **0,25–1 CU** com scale-to-zero de 5 min. Os crons, a ingestão diária e o tráfego mantêm o compute ativo ~400–500 h/mês, com média de ~0,4–0,5 CU, o que dá **~200 CU-h ≈ US$ 21**.
- Crescimento: 0,5–2 CU, praticamente sempre ligado (~1 CU médio × 730 h ≈ **US$ 77**). Escala: 1–4 CU (~2,5 CU médio ≈ 1.825 CU-h ≈ **US$ 193**). Antes de subir o compute, use as leituras cacheadas (`unstable_cache`) e o pooler.
- A região do banco deve ser a mesma das Functions (iad1 ↔ AWS us-east-1).

## 4. Upstash Redis (Marketplace)

Fonte: [Upstash Redis pricing](https://upstash.com/pricing/redis) (conferido em 2026-10-03).

| Plano | Preço | Dados | Banda | Observação |
| --- | --- | --- | --- | --- |
| Free | US$ 0 | 256 MB | 10 GB | **500K comandos/mês** |
| Pay as You Go | **US$ 0,20 / 100K comandos** | até 100 GB (1º GB grátis, depois US$ 0,25/GB) | 200 GB grátis, depois US$ 0,03/GB | budget mensal opcional (rate limit ao atingir) |
| Fixed 250 MB | **US$ 10/mês** | 250 MB | 50 GB | sem cobrança por comando, 10k cmd/s |
| Fixed 1 GB | US$ 20/mês | 1 GB | 100 GB | idem |
| Prod Pack (add-on) | não consegui confirmar o preço na página | — | — | SLA, multi-zone, SOC 2, criptografia em repouso |

AUTH/PING/INFO etc. não contam. Com réplica de leitura global, cada escrita conta em dobro (o app não usa).

**Comandos por mês** (do `quota.ts`): ~4,5 comandos por request da API, mais o `flush-usage` diário (alguns `HGETALL`/`DEL`). O Free tem 100 requests no total por conta, então quase todo o volume vem de Pro/Business.

| Cenário | Requests da API/mês | Comandos/mês | Pay as You Go | Melhor opção |
| --- | --- | --- | --- | --- |
| Lançamento | 50 mil (~1 cliente Pro) | ~230 mil | US$ 0,45 | **Free (US$ 0)** |
| Crescimento | 1,5 milhão (~20 Pro + 1 Business usando ~60%) | ~6,8 milhões | US$ 13,50 | **Fixed 250 MB (US$ 10)** |
| Escala | 10 milhões | ~45 milhões | US$ 90 | **Fixed 250 MB (US$ 10)**: ~45M × ~0,4 KB ≈ 18 GB de banda, abaixo dos 50 GB. Fixed 1 GB (US$ 20) se a banda passar de 50 GB |

Recomendação: começar no **Free**. Quando passar de ~400K comandos/mês, ir direto para o **Fixed 250 MB** (a partir de ~5M comandos ele sai mais barato que o Pay as You Go). Como a cota é *fail closed* (503 se o Redis falhar), não deixe o Free estourar em produção.

## 5. Demais serviços

| Serviço | Preço (fonte, 2026-10-03) | Uso no Hexascore |
| --- | --- | --- |
| **Resend** | Free: 3.000 e-mails/mês, **máx. 100/dia**, 3 domínios. Pro: US$ 20/mês (50k) ou US$ 35 (100k), excedente US$ 0,90/1K, sem limite diário ([resend.com/pricing](https://resend.com/pricing)) | verificação, reset, 2FA, lembretes de cobrança. No lançamento cabe no Free, mas os 100/dia podem travar um pico de cadastros: vá para o Pro ao divulgar |
| **Domínio `hexascore.app`** | Registrar da Vercel: compra US$ 9,99, **renovação US$ 15/ano** (API `get_tld_price`, 2026-10-03). O Pro inclui **1 domínio grátis no 1º ano** em TLD elegível, `.app` incluso ([Pro plan](https://vercel.com/docs/plans/pro-plan)) | US$ 1,25/mês amortizado a partir do 2º ano |
| **GitHub Actions** | Repositório público com runners padrão: **grátis**. Privado: GitHub Free tem 2.000 min/mês (Pro/Team: 3.000), Linux 2-core US$ 0,006/min ([billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions), [rates](https://docs.github.com/en/billing/reference/actions-minute-multipliers)) | Repositório **privado**. Estimativa: incremental 30 × ~45 min + semanal ~4 × ~200 + CI ~40 PRs × ~20 + QA ~8 × ~25 + CodeQL ≈ **3.300 min/mês → ~1.300 min excedentes ≈ US$ 8**. Cada bootstrap (6 jobs de até 6 h) pode somar até ~2.000 min (≈ US$ 12) |
| **Stripe (conta BR)** | Cartão nacional **3,99% + R$ 0,39**; cartão internacional **+2%**; Stripe Billing (assinaturas) **0,7% do volume** ([stripe.com/br/pricing](https://stripe.com/br/pricing)) | Pro R$ 29: ~R$ 1,75 (6,0%). Business R$ 149: ~R$ 7,38 (5,0%). Cobranças em USD/EUR caem na tarifa internacional (+2%). Não confirmei a tarifa de conversão de moeda |
| **AbacatePay (PIX)** | **R$ 0,80 por PIX**, sem mensalidade; boleto R$ 2,50 ([abacatepay.com](https://www.abacatepay.com/), página inicial) | Pro mensal R$ 29: 2,8%; anual R$ 290: 0,28%. Não encontrei uma página de preços dedicada nem tarifas de saque |

As taxas de pagamento são proporcionais à receita e ficam **fora** dos totais fixos abaixo.

## 6. Cenários

Premissas (todas ajustáveis):

| | Lançamento / baixo tráfego | Crescimento | Escala |
| --- | --- | --- | --- |
| Page views/mês | 30 mil | 300 mil | 3 milhões |
| Views de capa/mês | 100 mil | 1 milhão | 10 milhões |
| Requests da API/mês | 50 mil | 1,5 milhão | 10 milhões |
| Capas no Blob | 100k × 60 KB (6 GB) | 100k × 60 KB | 1M × 60 KB + miniatura (65 GB) |
| E-mails/mês | ~1,5 mil | ~15 mil | ~50–100 mil |
| Banco | 5 GB, ~200 CU-h | 5 GB, ~730 CU-h | 10 GB, ~1.825 CU-h |

Hipóteses de compute (iad1): por page view, 2 invocações (HTML + RSC), 50 ms de CPU ativa e ~0,2 GB-s de memória provisionada (já dividida pela concorrência do Fluid); por request da API, 20 ms e 0,08 GB-s; ~5 leituras de Runtime Cache por page view; ~300 KB de FDT por page view; ~10 CDN requests por page view + 1 por view de capa; 1 evento de Web Analytics por page view; ~80 builds/mês em Elastic (~US$ 6).

| Item (USD/mês) | Lançamento | Crescimento | Escala |
| --- | --- | --- | --- |
| Vercel Pro (taxa) | 20,00 | 20,00 | 20,00 |
| Functions (invocações + CPU + memória) | 0,26 | 4,10 | 33,27 |
| Runtime Cache | 0,56 | 1,10 | 6,50 |
| Web Analytics | 0,90 | 9,00 | 90,00 |
| Builds (Elastic) | ~6,00 | ~6,00 | ~8,00 |
| Blob (armazenamento + FOT/ops de MISS) | 0,15 | 0,55 | 5,50 |
| Uso Vercel − crédito de US$ 20 | 0 (uso ~8 < 20) | ~0,75 | ~123,27 |
| Flat Rate CDN (tier) | 0 (0,4M req, ~15 GB) | 20,00 (4M req) | 100,00 (40M req, ~1,5 TB) |
| **Subtotal Vercel** | **20,00** | **~40,75** | **~243,27** |
| Neon Launch (compute + storage + histórico) | ~23,00 | ~80,00 | ~199,00 |
| Upstash | 0 (Free) | 10,00 (Fixed 250 MB) | 10,00 (Fixed 250 MB) |
| Resend | 0 (Free) | 20,00 (Pro 50k) | 20,00–35,00 (Pro) |
| Domínio (renovação amortizada) | 1,25 (0 no 1º ano) | 1,25 | 1,25 |
| GitHub Actions (repositório privado) | ~8,00 | ~8,00 | ~10,00 |
| **Total USD/mês** | **~US$ 52** | **~US$ 160** | **~US$ 484–499** |
| **Total BRL/mês (× 5,2238)** | **~R$ 273** | **~R$ 836** | **~R$ 2.526–2.604** |

Leitura rápida: no lançamento o custo é praticamente **Vercel Pro + Neon Launch**. O lançamento empata com ~10 assinaturas Pro (R$ 29 − ~R$ 1,75 de taxa). Em escala, os maiores itens são **Neon compute**, o **tier da CDN** e o **Web Analytics** (US$ 90 com 3M de page views, porque o Pro não inclui eventos).

## 7. Controle de custos

1. **Spend Management** (Settings → Billing → Spend Management, papel Owner/Billing): defina um teto, notificações (web/e-mail/SMS) e um webhook. "Pause Production Deployments" derruba **todos** os projetos com 503 até alguém retomar manualmente, e a checagem pode atrasar alguns minutos, então deixe o teto abaixo do máximo aceitável ([docs](https://vercel.com/docs/spend-management)).
2. **Flat Rate CDN:** confira se está ligado e acompanhe o uso diário com `vercel usage`. Um tier pago (US$ 20 para 10M req) sai mais barato que o on-demand a partir de ~6–7M requests/mês com banda relevante.
3. **Capas no Blob:** grave com `cacheControlMaxAge` longo (ex.: 1 ano), porque o path `covers/<item_id>` é imutável. Trocar uma capa = nova chave (ex.: hash no nome), nunca sobrescrever, assim o HIT ratio fica próximo de 100% e não há Simple Operations nem Fast Origin Transfer.
4. **Tamanho das imagens:** converta para WebP no `covers.py` (alvo ≤ 60 KB por capa, largura ~600 px) e gere a miniatura de 5 KB para grades e autocomplete. Isso evita ativar o Image Optimization (US$ 0,05/1K transformações) e reduz a banda em ~90% nas listagens. Use `loading="lazy"` e `width`/`height` nas `<img>`.
5. **Uploads idempotentes:** não reenviar capas já gravadas (cada `put` é uma Advanced Operation de US$ 5/1M) e respeitar o limite de 75 put/s.
6. **Functions:** as páginas são dinâmicas por causa do nonce, então mantenha as leituras de catálogo no `unstable_cache` (US$ 0,40/1M leituras) em vez de bater no Neon. Não mude a região para gru1 sem motivo (~1,7× mais cara em CPU e memória).
7. **Web Analytics:** é o único item que cresce linearmente com page views sem franquia. Em escala, avalie limitar a coleta a páginas-chave ou trocar por outra ferramenta. O Speed Insights fica no free tier (sem cobrança). Não ative o Speed Insights Plus sem necessidade.
8. **Builds:** troque Elastic por **Standard** em Settings → Build and Deployment (Standard não é cobrado no Pro sem on-demand concurrency) e use `ignoreCommand` para pular builds de mudanças só em `docs/` ou `webscrappings/`.
9. **Neon:** scale-to-zero ligado enquanto o tráfego for baixo, autoscaling com teto, poucas branches de preview (cada uma acima de 10 custa US$ 1,50/mês) e limpeza das branches de PRs fechados.
10. **Upstash:** Free → Fixed 250 MB (sem cobrança por comando). Se usar Pay as You Go, configure o budget. Mantenha os `EXPIRE` já presentes no `quota.ts`.
11. **GitHub Actions:** o maior consumo é a ingestão. Tornar o repositório público zera o custo de minutos. No privado, reduza o semanal ou o `timeout-minutes` e evite rodar bootstrap à toa.

## Preços não confirmados

- Se os planos e preços do Neon e da Upstash **dentro do Vercel Marketplace** são idênticos aos das páginas próprias (a Vercel mostra só os nomes dos planos e "starting at $0").
- Se o crédito de US$ 20 do Pro abate produtos do Marketplace ou o tier pago do Flat Rate CDN (assumido que não).
- Se o Flat Rate CDN já vem ligado no time ou exige ativação.
- Preço do Upstash Prod Pack, tarifa de conversão de moeda da Stripe BR e tarifas de saque da AbacatePay.
- Tamanho real de `source_records` (medir depois do bootstrap).

# DOC-i18n-legal Proposed wording for legal keys (Q-D-007), no catalogue edit
status: concluida (proposal only; owner approval and lawyer read required before any edit)
branch: none (report only); main @ 13a6e08
files: .orchestration/reports/DOC-i18n-legal.md
acceptance:
- `bun -e` read of `resources/lang/{en,pt-BR,es}.json` → 8 keys printed in all 3 locales (current text, quoted below)
- `grep -rn "passkey" resources/lang/` → 0 matches (sources: command run 2026-10-31)
- `grep -rn "terms.sections.5\|policy.sections.1.body" --include=*.go --include=*.tsx --include=*.ts` → only `.orchestration` and `reports` hits, no code reference
- decision sources: `docs/brainstorming.md` line 47 (D-06 Q4, amended 2026-10-29: passkeys no longer deferred); line 65 (Q13: PIX in second release); `docs/orchestration/QUESTIONS.md` line 42 (passkeys come back), line 84 (social layer is Phase 7)
validation: lint n/a (no edit), typecheck n/a, tests n/a, browser n/a
changelog: none
roadmap: no box
risks / follow-ups:
- Q-D-007 premise is stale: current catalogues have no passkey mention. Earlier report `E-015.md` quoted older wording (e.g. "Card and PIX payment data ... Stripe and AbacatePay") that is not in the files now. Use the current text below as the base.
- Passkeys are owner-approved to return (QUESTIONS.md l.42), so do not add a "no passkeys" clause; when ADOPT-VF-001 merges, add passkey wording to `policy.sections.1.body.0` and `policy.sections.7.body.0` in all 3 locales.
- Follows, PIX/AbacatePay and credits stay deferred (D-06, Q13, D-23), so removing them from the copy is consistent with the decisions. Re-add in the release that ships each one.
- Legal copy needs owner and lawyer review; pt-BR and es wording below is my draft, for a text-editor pass and owner sign-off.

## Current text (en, read from catalogue)

- `policy.sections.1.body.1`: "Activity: your ratings (six categories, 0–100), item status, lists, follows and profile privacy choice."
- `policy.sections.1.body.4`: "API and billing: API key metadata (we store only a hash of the key), request counts, plan, credits and invoices. Card payments are handled by Stripe; we never see or store full card numbers."
- `terms.sections.5.body.0`: "... By PIX, through AbacatePay (Brazilian reais only), you buy 1 month or 12 months ... PIX periods do not renew automatically ... Prices, currency and taxes are shown before you pay."
- `terms.sections.5.body.1`: "Extra credits can only be bought with an active Pro or Business plan; they do not expire and are used after the monthly quota."
- `terms.sections.5.body.2`: "... If you are charged twice for the same period (for example, a PIX payment while a card subscription is active), we refund the duplicate payment."

## Proposed wording

Keys `policy.sections.1.body.0`, `policy.sections.6.body.1`, `policy.sections.7.body.0` need no change now (no PIX, credits, follows or passkey text).

### en

- `policy.sections.1.body.1`: "Activity: your ratings (six categories, 0–100), item status, lists and profile privacy choice."
- `policy.sections.1.body.4`: "API and billing: API key metadata (we store only a hash of the key), request counts, plan and invoices. Card payments are handled by Stripe; we never see or store full card numbers."
- `terms.sections.5.body.0`: "Paid plans are billed in advance. By card, through Stripe, plans are monthly subscriptions that renew automatically until cancelled; you can cancel or change your card in the billing portal. Prices, currency and taxes are shown before you pay."
- `terms.sections.5.body.1`: delete this key in all 3 catalogues; renumber `terms.sections.5.body.2` to `body.1` (parity test needs the same key set).
- `terms.sections.5.body.1` (new, was body.2): "You can cancel at any time and keep access until the end of the paid period. Where the law grants you a right of withdrawal (for example, 7 days under art. 49 of the Brazilian Consumer Code), you can exercise it through the contact form and receive a refund as the law provides. If you are charged twice for the same period, we refund the duplicate payment."

### pt-BR

- `policy.sections.1.body.1`: "Atividade: suas avaliações (seis categorias, 0–100), status dos itens, listas e a escolha de perfil privado."
- `policy.sections.1.body.4`: "API e cobrança: metadados das chaves de API (guardamos só um hash da chave), contagem de requisições, plano e faturas. Pagamentos com cartão são tratados pela Stripe; nunca vemos nem guardamos o número completo do cartão."
- `terms.sections.5.body.0`: "Os planos pagos são cobrados antecipadamente. No cartão, pela Stripe, os planos são assinaturas mensais que renovam automaticamente até serem canceladas; você pode cancelar ou trocar o cartão no portal de cobrança. Preços, moeda e impostos aparecem antes do pagamento."
- `terms.sections.5.body.1`: delete; renumber body.2 to body.1 as in en.
- `terms.sections.5.body.1` (was body.2): "Você pode cancelar a qualquer momento e mantém o acesso até o fim do período pago. Quando a lei garante o direito de arrependimento (por exemplo, 7 dias pelo art. 49 do Código de Defesa do Consumidor), você pode exercê-lo pelo formulário de contato e receber o reembolso previsto em lei. Se você for cobrado duas vezes pelo mesmo período, devolvemos o pagamento em duplicidade."

### es

- `policy.sections.1.body.1`: "Actividad: tus puntuaciones (seis categorías, 0–100), estado de las obras, listas y la elección de perfil privado."
- `policy.sections.1.body.4`: "API y facturación: metadatos de las claves de API (guardamos solo un hash de la clave), número de solicitudes, plan y facturas. Los pagos con tarjeta los trata Stripe; nunca vemos ni guardamos el número completo de la tarjeta."
- `terms.sections.5.body.0`: "Los planes de pago se cobran por adelantado. Con tarjeta, a través de Stripe, los planes son suscripciones mensuales que se renuevan automáticamente hasta que las canceles; puedes cancelar o cambiar la tarjeta en el portal de facturación. Los precios, la moneda y los impuestos se muestran antes de pagar."
- `terms.sections.5.body.1`: delete; renumber body.2 to body.1 as in en.
- `terms.sections.5.body.1` (was body.2): "Puedes cancelar en cualquier momento y mantienes el acceso hasta el final del periodo pagado. Cuando la ley te concede un derecho de desistimiento (por ejemplo, 7 días según el art. 49 del Código de Defensa del Consumidor de Brasil), puedes ejercerlo mediante el formulario de contacto y recibir el reembolso que establezca la ley. Si se te cobra dos veces por el mismo periodo, devolvemos el pago duplicado."

## Review (especialista, 2026-10-31)
verdict: APROVADO (proposal only; no catalogue edit until owner and lawyer sign off, Q-D-007)
re-run: `bun -e` read of resources/lang/en.json -> `policy.sections.1.body.1` and `terms.sections.5.body.0..2` match the quoted text (follows, PIX/AbacatePay, credits present)
checked: decisions cited (D-06 Q4 amended, Q13, D-23); key renumbering keeps the parity set; pt-BR/es drafts read, faithful to en
notes: 1. ADOPT-VF-001 (passkeys) is now merged on main and `passkey` strings exist in all 3 catalogues; the report's "when ADOPT-VF-001 merges" follow-up is due: `policy.sections.1.body.0` lists only "email and password, or Google" as sign-in methods, add passkeys in the same owner-approved edit. 2. The proposed `terms.sections.5.body.0` says card plans are monthly only; confirm with the Stripe price set (Q-H31) before applying.

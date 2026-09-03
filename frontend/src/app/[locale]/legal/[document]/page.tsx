import { hasLocale } from "next-intl";
import { notFound } from "next/navigation";

import { routing, type AppLocale } from "@/i18n/routing";

const documents = [
  "terms",
  "privacy",
  "refund-policy",
  "payment-policy",
] as const;
type Document = (typeof documents)[number];

type Policy = {
  title: string;
  updated: string;
  intro: string;
  sections: Array<{ title: string; body: string }>;
};

const policy: Record<AppLocale, Record<Document, Policy>> = {
  "pt-PT": {
    terms: {
      title: "Termos da Vila",
      updated: "Atualizado em 2 de setembro de 2026",
      intro:
        "A Vila é um marketplace que aproxima clientes e prestadores independentes. Ao utilizar a plataforma, aceita estes termos e as políticas de pagamento e reembolso.",
      sections: [
        {
          title: "Papel da plataforma",
          body: "A Vila facilita descoberta, comunicação, reservas e pagamentos. O prestador continua responsável pela descrição, legalidade, qualidade e execução do serviço. A Vila não é o empregador nem o executante do serviço.",
        },
        {
          title: "Contas e anúncios",
          body: "As informações devem ser verdadeiras, atuais e não enganosas. Anúncios são sujeitos a revisão e podem ser rejeitados, suspensos ou removidos por segurança, fraude, ilegalidade ou incumprimento.",
        },
        {
          title: "Reservas",
          body: "Preço, data, local e âmbito devem ser confirmados antes do pagamento. Alterações materiais exigem acordo entre cliente e prestador. As ações ficam registadas para segurança e resolução de litígios.",
        },
        {
          title: "Lei e contacto",
          body: "Aplicam-se os direitos imperativos do consumidor e a legislação portuguesa e europeia aplicável. Questões podem ser enviadas para source.sensei1205@gmail.com. A identificação legal e fiscal completa do operador deve constar do aviso comercial antes da ativação pública de pagamentos.",
        },
      ],
    },
    privacy: {
      title: "Privacidade",
      updated: "Atualizado em 2 de setembro de 2026",
      intro:
        "A Vila minimiza os dados que recolhe e separa identidade, contactos privados, pagamentos e informação pública.",
      sections: [
        {
          title: "Dados tratados",
          body: "Tratamos identidade de conta, perfil, anúncios, conversas, reservas, eventos de moderação e referências de pagamento. Não armazenamos números de cartão nem dados bancários; a Clerk processa identidade e a Stripe processa pagamentos e verificação de recebimentos.",
        },
        {
          title: "Finalidades e conservação",
          body: "Os dados servem para prestar o serviço, prevenir fraude, cumprir obrigações legais, resolver litígios e manter registos financeiros. A conservação deve limitar-se ao período necessário e aos prazos legais aplicáveis.",
        },
        {
          title: "Direitos",
          body: "Pode pedir acesso, correção, exportação, oposição ou eliminação quando legalmente possível através de source.sensei1205@gmail.com. Dados exigidos por obrigações financeiras, fiscais ou de segurança podem ter de ser conservados.",
        },
      ],
    },
    "refund-policy": {
      title: "Cancelamentos e reembolsos",
      updated: "Atualizado em 2 de setembro de 2026",
      intro:
        "Os pedidos são avaliados com base no estado da reserva, execução do serviço, acordo entre as partes e direitos legais aplicáveis.",
      sections: [
        {
          title: "Antes do serviço",
          body: "Um cancelamento antes do início pode originar reembolso total ou parcial conforme custos já incorridos e termos claramente aceites na reserva. Nenhuma regra reduz direitos imperativos do consumidor.",
        },
        {
          title: "Serviço iniciado ou concluído",
          body: "Reembolsos após o início exigem análise do trabalho executado, evidência e acordo ou decisão administrativa. Serviços personalizados ou iniciados com consentimento podem ter regras legais específicas.",
        },
        {
          title: "Processamento",
          body: "Reembolsos aprovados são enviados à Stripe para o método de pagamento original. O prazo bancário depende do método e da instituição. Fraude, abuso e chargebacks podem suspender pagamentos e recebimentos durante a investigação.",
        },
      ],
    },
    "payment-policy": {
      title: "Pagamentos e recebimentos",
      updated: "Atualizado em 2 de setembro de 2026",
      intro:
        "A Vila utiliza Checkout e Connect alojados pela Stripe para que dados financeiros sensíveis não passem pelos formulários da Vila.",
      sections: [
        {
          title: "Preço e taxas",
          body: "Antes de continuar para a Stripe, o cliente vê o valor total, a taxa da plataforma e o valor previsto para o prestador. Os valores são calculados no servidor em euros e não podem ser escolhidos pelo navegador.",
        },
        {
          title: "Métodos e impostos",
          body: "Os métodos disponíveis, incluindo MB WAY quando elegível, dependem da conta Stripe, país, moeda, montante e configuração. A Stripe pode recolher morada e identificação fiscal e gerar documentação de pagamento; o cumprimento e entrega de declarações fiscais permanecem responsabilidade das partes aplicáveis.",
        },
        {
          title: "Recebimentos e litígios",
          body: "Prestadores concluem verificação diretamente na Stripe. Pagamentos, reembolsos, disputas e chargebacks são sincronizados por webhooks assinados. A Vila pode suspender recebimentos enquanto uma disputa está aberta ou quando a Stripe limita a conta.",
        },
      ],
    },
  },
  en: {} as Record<Document, Policy>,
  es: {} as Record<Document, Policy>,
};
policy.en = translateEnglish();
policy.es = translateSpanish();

export function generateStaticParams() {
  return routing.locales.flatMap((locale) =>
    documents.map((document) => ({ locale, document })),
  );
}

export default async function LegalPage({
  params,
}: PageProps<"/[locale]/legal/[document]">) {
  const { locale, document } = await params;
  if (
    !hasLocale(routing.locales, locale) ||
    !documents.includes(document as Document)
  )
    notFound();
  const value = policy[locale][document as Document];
  return (
    <main className="market-page px-4 py-10 sm:px-6 sm:py-16">
      <article className="market-panel mx-auto max-w-4xl p-6 sm:p-10">
        <p className="market-kicker">Vila</p>
        <h1 className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl">
          {value.title}
        </h1>
        <p className="mt-3 text-sm text-muted">{value.updated}</p>
        <p className="mt-6 text-lg leading-8 text-muted">{value.intro}</p>
        <div className="mt-10 grid gap-8">
          {value.sections.map((section) => (
            <section key={section.title}>
              <h2 className="text-2xl font-semibold">{section.title}</h2>
              <p className="mt-3 leading-7 text-muted">{section.body}</p>
            </section>
          ))}
        </div>
      </article>
    </main>
  );
}

function translateEnglish(): Record<Document, Policy> {
  const updated = "Updated 2 September 2026";
  return {
    terms: {
      title: "Vila terms",
      updated,
      intro:
        "Vila is a marketplace connecting customers with independent providers. By using it, you accept these terms and the payment and refund policies.",
      sections: [
        {
          title: "Platform role",
          body: "Vila facilitates discovery, communication, bookings, and payments. Providers remain responsible for the description, legality, quality, and delivery of their services. Vila is neither the provider nor their employer.",
        },
        {
          title: "Accounts and listings",
          body: "Information must be accurate, current, and not misleading. Listings are reviewed and may be rejected, suspended, or removed for safety, fraud, illegality, or breach.",
        },
        {
          title: "Bookings",
          body: "Price, date, private location, and scope must be confirmed before payment. Material changes require agreement. Actions are recorded for safety and dispute resolution.",
        },
        {
          title: "Law and contact",
          body: "Mandatory consumer rights and applicable Portuguese and European law remain in force. Contact source.sensei1205@gmail.com. The operator's complete legal and tax identity must appear in the commercial notice before public payment activation.",
        },
      ],
    },
    privacy: {
      title: "Privacy",
      updated,
      intro:
        "Vila minimizes data and separates identity, private contact details, payments, and public information.",
      sections: [
        {
          title: "Data processed",
          body: "We process account identity, profiles, listings, conversations, bookings, moderation events, and payment references. We do not store card numbers or bank details; Clerk processes identity and Stripe processes payments and payout verification.",
        },
        {
          title: "Purpose and retention",
          body: "Data is used to provide the service, prevent fraud, meet legal obligations, resolve disputes, and maintain financial records. Retention is limited to necessity and applicable legal periods.",
        },
        {
          title: "Your rights",
          body: "Request access, correction, export, objection, or deletion where legally available through source.sensei1205@gmail.com. Financial, tax, fraud-prevention, and security records may need to be retained.",
        },
      ],
    },
    "refund-policy": {
      title: "Cancellations and refunds",
      updated,
      intro:
        "Requests are assessed from booking state, work performed, agreement between the parties, and applicable statutory rights.",
      sections: [
        {
          title: "Before work starts",
          body: "Cancellation before work begins may qualify for a full or partial refund according to costs already incurred and terms clearly accepted in the booking. Mandatory consumer rights are not reduced.",
        },
        {
          title: "Started or completed work",
          body: "Refunds after work starts require review of delivery, evidence, and agreement or an administrative decision. Customized services or services begun with consent may have specific legal rules.",
        },
        {
          title: "Processing",
          body: "Approved refunds are submitted to Stripe for the original payment method. Bank timing depends on the method and institution. Fraud, abuse, and chargebacks can suspend payments and payouts during investigation.",
        },
      ],
    },
    "payment-policy": {
      title: "Payments and payouts",
      updated,
      intro:
        "Vila uses Stripe-hosted Checkout and Connect so sensitive financial details never pass through Vila forms.",
      sections: [
        {
          title: "Price and fees",
          body: "Before continuing to Stripe, customers see the total, platform fee, and expected provider payout. Values are calculated server-side in euros and cannot be selected by the browser.",
        },
        {
          title: "Methods and tax",
          body: "Available methods, including MB WAY when eligible, depend on Stripe account, country, currency, amount, and settings. Stripe can collect addresses and tax IDs and create payment documents; applicable filing and tax obligations remain with the relevant parties.",
        },
        {
          title: "Payouts and disputes",
          body: "Providers complete verification directly with Stripe. Payments, refunds, disputes, and chargebacks synchronize through signed webhooks. Vila may suspend payouts while a dispute is open or Stripe restricts an account.",
        },
      ],
    },
  };
}

function translateSpanish(): Record<Document, Policy> {
  const updated = "Actualizado el 2 de septiembre de 2026";
  return {
    terms: {
      title: "Términos de Vila",
      updated,
      intro:
        "Vila es un marketplace que conecta clientes con profesionales independientes. Al utilizarlo, aceptas estos términos y las políticas de pago y reembolso.",
      sections: [
        {
          title: "Función de la plataforma",
          body: "Vila facilita el descubrimiento, la comunicación, las reservas y los pagos. El profesional sigue siendo responsable de la descripción, legalidad, calidad y prestación del servicio. Vila no es el profesional ni su empleador.",
        },
        {
          title: "Cuentas y anuncios",
          body: "La información debe ser veraz, actual y no engañosa. Los anuncios se revisan y pueden rechazarse, suspenderse o eliminarse por seguridad, fraude, ilegalidad o incumplimiento.",
        },
        {
          title: "Reservas",
          body: "El precio, la fecha, la ubicación privada y el alcance deben confirmarse antes del pago. Los cambios importantes requieren acuerdo. Las acciones quedan registradas para seguridad y resolución de conflictos.",
        },
        {
          title: "Ley y contacto",
          body: "Siguen vigentes los derechos obligatorios de los consumidores y la legislación portuguesa y europea aplicable. Contacto: source.sensei1205@gmail.com. La identidad legal y fiscal completa del operador debe aparecer en el aviso comercial antes de activar pagos públicos.",
        },
      ],
    },
    privacy: {
      title: "Privacidad",
      updated,
      intro:
        "Vila minimiza los datos y separa identidad, contactos privados, pagos e información pública.",
      sections: [
        {
          title: "Datos tratados",
          body: "Tratamos identidad de cuenta, perfiles, anuncios, conversaciones, reservas, eventos de moderación y referencias de pago. No almacenamos números de tarjeta ni datos bancarios; Clerk trata la identidad y Stripe los pagos y la verificación de cobros.",
        },
        {
          title: "Finalidad y conservación",
          body: "Los datos se usan para prestar el servicio, prevenir fraude, cumplir obligaciones legales, resolver conflictos y conservar registros financieros. La conservación se limita a lo necesario y a los plazos legales aplicables.",
        },
        {
          title: "Tus derechos",
          body: "Puedes solicitar acceso, rectificación, exportación, oposición o supresión cuando sea legalmente posible mediante source.sensei1205@gmail.com. Puede ser necesario conservar datos financieros, fiscales, de prevención de fraude y seguridad.",
        },
      ],
    },
    "refund-policy": {
      title: "Cancelaciones y reembolsos",
      updated,
      intro:
        "Las solicitudes se evalúan según el estado de la reserva, el trabajo realizado, el acuerdo entre las partes y los derechos legales aplicables.",
      sections: [
        {
          title: "Antes de comenzar",
          body: "Una cancelación antes del inicio puede dar derecho a un reembolso total o parcial según los costes ya incurridos y las condiciones aceptadas claramente en la reserva. No se reducen los derechos obligatorios del consumidor.",
        },
        {
          title: "Servicio iniciado o finalizado",
          body: "Los reembolsos tras el inicio requieren revisar la prestación, las pruebas y el acuerdo o una decisión administrativa. Los servicios personalizados o iniciados con consentimiento pueden tener reglas legales específicas.",
        },
        {
          title: "Procesamiento",
          body: "Los reembolsos aprobados se envían a Stripe para el método de pago original. El plazo bancario depende del método y la entidad. El fraude, abuso y los contracargos pueden suspender pagos y cobros durante la investigación.",
        },
      ],
    },
    "payment-policy": {
      title: "Pagos y cobros",
      updated,
      intro:
        "Vila utiliza Checkout y Connect alojados por Stripe para que los datos financieros sensibles no pasen por los formularios de Vila.",
      sections: [
        {
          title: "Precio y comisiones",
          body: "Antes de continuar a Stripe, el cliente ve el total, la comisión de la plataforma y el cobro previsto del profesional. Los importes se calculan en el servidor en euros y el navegador no puede elegirlos.",
        },
        {
          title: "Métodos e impuestos",
          body: "Los métodos disponibles, incluido MB WAY cuando corresponda, dependen de la cuenta Stripe, país, moneda, importe y configuración. Stripe puede recoger dirección e identificación fiscal y crear documentos de pago; las obligaciones fiscales aplicables siguen correspondiendo a las partes pertinentes.",
        },
        {
          title: "Cobros y conflictos",
          body: "Los profesionales completan la verificación directamente con Stripe. Pagos, reembolsos, disputas y contracargos se sincronizan mediante webhooks firmados. Vila puede suspender cobros mientras exista una disputa o Stripe restrinja una cuenta.",
        },
      ],
    },
  };
}

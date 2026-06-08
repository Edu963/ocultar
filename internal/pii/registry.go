package pii

import "regexp"

type ValidationMethod string

const (
	ValNone  ValidationMethod = "NONE"
	ValLuhn  ValidationMethod = "LUHN"
	ValMod97 ValidationMethod = "MOD97"
	ValESDNI ValidationMethod = "ES_DNI"
	ValITCF  ValidationMethod = "IT_CF"
	ValNLBSN ValidationMethod = "NL_BSN"
	ValPLPSL ValidationMethod = "PL_PESEL"
	ValDESTID ValidationMethod = "DE_STID"
	ValDKCPR  ValidationMethod = "DK_CPR"
	ValFIHETU ValidationMethod = "FI_HETU"
	ValSEPIN  ValidationMethod = "SE_PIN"
	ValBRCPF  ValidationMethod = "BR_CPF"
	ValCLRUT  ValidationMethod = "CL_RUT"
	ValIndiaAadhaar ValidationMethod = "INDIA_AADHAAR"
	ValSingaporeID  ValidationMethod = "SG_ID"
	ValESCIF        ValidationMethod = "ES_CIF"
)

type EntityDef struct {
	Type          string
	Pattern       *regexp.Regexp
	Validator     ValidationMethod
	MinLength     int
	Normalization bool // If true, caller should strip spaces/dashes before validation
	CaptureGroup  int  // If > 0, only this capture group is tokenized
}

var Registry = []EntityDef{
	// Tier 0 Cloud Secrets (RAM Rule Patch)
	{Type: "AWS_KEY", Pattern: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), Validator: ValNone, MinLength: 20, Normalization: false},
	{Type: "AWS_SECRET", Pattern: regexp.MustCompile(`\b[0-9a-zA-Z/+]{40}\b`), Validator: ValNone, MinLength: 40, Normalization: false},
	{Type: "GCP_SERVICE_ACCOUNT", Pattern: regexp.MustCompile(`(?i)\b[a-z0-9-]+@[a-z0-9-]+\.iam\.gserviceaccount\.com\b`), Validator: ValNone, MinLength: 15, Normalization: false},
	{Type: "IP_ADDRESS", Pattern: regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`), Validator: ValNone, MinLength: 7, Normalization: false},
	{Type: "PROJECT_CODE", Pattern: regexp.MustCompile(`(?i)\bProject (Phoenix|Ouroboros|Titan)\b`), Validator: ValNone, MinLength: 10, Normalization: false},
	// Internal cost-center / GL codes: e.g. 6837-CORP-891, 1076-EMEA-824, 5866-US-432
	{Type: "COST_CENTER", Pattern: regexp.MustCompile(`\b\d{3,6}-[A-Z]{2,8}-\d{3,4}\b`), Validator: ValNone, MinLength: 8, Normalization: false},

	// Financial
	// Primary IBAN — full match with Mod97 checksum (complete IBANs only)
	{Type: "IBAN", Pattern: regexp.MustCompile(`(?i)\b[A-Z]{2}[0-9]{2}(?:[A-Z0-9]{11,30}|(?:[\s-][A-Z0-9]{4}){2,7}[\s-]?[A-Z0-9]{0,3})\b`), Validator: ValMod97, MinLength: 15, Normalization: true, CaptureGroup: 0},
	// Context-gated partial IBAN — catches IBANs already redacted in source with ***
	// (e.g. "FR76 1390 6001 0083 1172 45******"). No Mod97: truncated IBANs cannot pass checksum.
	{Type: "IBAN", Pattern: regexp.MustCompile(`(?i)(?:\bcompte\s+bancaire\b|\biban\b|\bvirement\b|\bpr[eé]lev[eé]\s+sur\b)[^A-Z0-9]{0,30}([A-Z]{2}[0-9]{2}(?:[\s-]?[A-Z0-9]{4}){2,6}(?:[\s-]?[A-Z0-9]{1,4})?(?:\s*\*+)?)`), Validator: ValNone, MinLength: 10, Normalization: false, CaptureGroup: 1},
	{Type: "CREDIT_CARD", Pattern: regexp.MustCompile(`\b(?:4[0-9\s-]{12,19}|5[1-5][0-9\s-]{14,19}|6(?:011|5[0-9]{2})[0-9\s-]{12,19}|3[47][0-9\s-]{13,19}|3(?:0[0-5]|[68][0-9])[0-9\s-]{11,19}|(?:2131|1800|35\d{3})[0-9\s-]{11,19})\b`), Validator: ValLuhn, MinLength: 13, Normalization: true, CaptureGroup: 0},
	{Type: "BIC", Pattern: regexp.MustCompile(`\b[A-Z]{6}[A-Z0-9]{2}(?:[A-Z0-9]{3})?\b`), Validator: ValNone, MinLength: 8, Normalization: false, CaptureGroup: 0},

	// EU/UK VAT (Generic EU VAT)
	{Type: "EU_VAT", Pattern: regexp.MustCompile(`(?i)\b(?:ATU[0-9]{8}|BE0[0-9]{9}|BG[0-9]{9,10}|CY[0-9]{8}[A-Z]|CZ[0-9]{8,10}|DE[0-9]{9}|DK[0-9]{8}|EE[0-9]{9}|EL[0-9]{9}|ES[A-Z0-9][0-9]{7}[A-Z0-9]|FI[0-9]{8}|FR[A-Z0-9]{2}[0-9]{9}|HR[0-9]{11}|HU[0-9]{8}|IE[0-9][A-Z0-9+*][0-9]{5}[A-Z]|IT[0-9]{11}|LT[0-9]{9,12}|LU[0-9]{8}|LV[0-9]{11}|MT[0-9]{8}|NL[0-9]{9}B[0-9]{2}|PL[0-9]{10}|PT[0-9]{9}|RO[0-9]{2,10}|SE[0-9]{12}|SI[0-9]{8}|SK[0-9]{10}|XI[0-9]{9})\b`), Validator: ValNone, MinLength: 6, Normalization: true},

	// France
	{Type: "FR_NIR", Pattern: regexp.MustCompile(`\b[12]\s*\d{2}\s*(?:0[1-9]|1[0-2])\s*(?:2[AB]|\d{2})\s*\d{3}\s*\d{3}\s*\d{2}\b`), Validator: ValNone, MinLength: 15, Normalization: true},
	{Type: "FRANCE_SIREN_NUMBER", Pattern: regexp.MustCompile(`\b[0-9]{3}\s*[0-9]{3}\s*[0-9]{3}\b`), Validator: ValLuhn, MinLength: 9, Normalization: true},
	{Type: "FRANCE_SIRET_NUMBER", Pattern: regexp.MustCompile(`\b[0-9]{3}\s*[0-9]{3}\s*[0-9]{3}\s*[0-9]{5}\b`), Validator: ValLuhn, MinLength: 14, Normalization: true},
	// French VAT — space-tolerant form catches "FR XX XXX XXX XXX" as printed on invoices.
	// EU_VAT pattern above only matches the compact no-space form "FRXXXXXXXXXX".
	{Type: "FRANCE_VAT", Pattern: regexp.MustCompile(`(?i)\bFR\s*[A-Z0-9]{2}\s*[0-9]{3}\s*[0-9]{3}\s*[0-9]{3}\b`), Validator: ValNone, MinLength: 11, Normalization: true},
	// French subscriber ID — context-gated to avoid NL_BSN false-positive on 8-9 digit IDs.
	// Note: no trailing \b after accented characters (Go regexp \b is ASCII-only).
	{Type: "FR_SUBSCRIBER_ID", Pattern: regexp.MustCompile(`(?i)(?:\bidentifiant\s+abonn[eé]|\bnum[eé]ro\s+abonn[eé]|\babonn[eé]\s*(?:id|n[o°]))[\s:]*([0-9]{7,15})\b`), Validator: ValNone, MinLength: 7, Normalization: false, CaptureGroup: 1},
	// French postal code — keyword-gated ("code postal", "CP", "cedex") to avoid false positives.
	{Type: "FR_POSTAL_CODE", Pattern: regexp.MustCompile(`(?i)(?:\bcode\s+postal\b|\bCP\b|\bc[eé]dex\b)[:\s]*([0-9]{5})\b`), Validator: ValNone, MinLength: 5, Normalization: false, CaptureGroup: 1},
	// French postal code — bare form, restricted to valid French department prefixes (01–95 metropolitan,
	// 971–976 overseas). Tighter than "any 5 digits"; catches footer codes like 75008, 38100, 75371
	// without a keyword anchor.
	{Type: "FR_POSTAL_CODE", Pattern: regexp.MustCompile(`\b(?:0[1-9]|[1-8][0-9]|9[0-5]|97[1-6])\d{3}\b`), Validator: ValNone, MinLength: 5, Normalization: false, CaptureGroup: 0},
	// French ISP fiber-optic reference — keyword-gated; prevents Tier A phone-number validation
	// from claiming the numeric reference as a PHONE (e.g. "Référence prise fibre : 0123456789012").
	{Type: "FIBER_REF", Pattern: regexp.MustCompile(`(?i)(?:\br[eé]f[eé]rence\s+(?:prise\s+)?(?:fibre|fiber)\b|\bprise\s+(?:fibre|fiber)\b)[:\s]*([A-Z0-9][A-Z0-9\-\.]{5,30})`), Validator: ValNone, MinLength: 5, Normalization: false, CaptureGroup: 1},

	// Spain
	{Type: "ES_DNI_NIE", Pattern: regexp.MustCompile(`(?i)\b[XYZ]?\d{7,8}[A-Z]\b`), Validator: ValESDNI, MinLength: 9, Normalization: true},
	{Type: "ES_CIF", Pattern: regexp.MustCompile(`(?i)\b[ABCDEFGHJNPQRSUVW]\d{7}[0-9A-J]\b`), Validator: ValESCIF, MinLength: 9, Normalization: true},

	// Germany
	{Type: "DE_STEUER_ID", Pattern: regexp.MustCompile(`\b\d{11}\b`), Validator: ValDESTID, MinLength: 11, Normalization: true},

	// Italy
	{Type: "IT_CODICE_FISCALE", Pattern: regexp.MustCompile(`(?i)\b[A-Z]{6}\d{2}[A-Z]\d{2}[A-Z]\d{3}[A-Z]\b`), Validator: ValITCF, MinLength: 16, Normalization: true},

	// Netherlands
	{Type: "NL_BSN", Pattern: regexp.MustCompile(`\b\d{8,9}\b`), Validator: ValNLBSN, MinLength: 8, Normalization: true},

	// UK
	{Type: "UK_NINO", Pattern: regexp.MustCompile(`(?i)\b[A-CEGHJ-PR-TW-Z][A-CEGHJ-NPR-TW-Z]\s*\d{2}\s*\d{2}\s*\d{2}\s*[A-D]\b`), Validator: ValNone, MinLength: 9, Normalization: true},
	{Type: "UK_NHS", Pattern: regexp.MustCompile(`\b\d{3}\s*\d{3}\s*\d{4}\b`), Validator: ValNone, MinLength: 10, Normalization: true},

	// Poland
	{Type: "PL_PESEL", Pattern: regexp.MustCompile(`\b\d{11}\b`), Validator: ValPLPSL, MinLength: 11, Normalization: true},

	// Nordics
	{Type: "FI_HETU", Pattern: regexp.MustCompile(`(?i)\b(?:0[1-9]|[12]\d|3[01])(?:0[1-9]|1[0-2])\d{2}[A+-]\d{3}[0-9A-FHJ-NPR-Y]\b`), Validator: ValFIHETU, MinLength: 11, Normalization: false},
	{Type: "SE_PIN", Pattern: regexp.MustCompile(`\b(18|19|20)?\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])[-+]?\d{4}\b`), Validator: ValSEPIN, MinLength: 10, Normalization: true},
	{Type: "DK_CPR", Pattern: regexp.MustCompile(`\b(?:0[1-9]|[12]\d|3[01])(?:0[1-9]|1[0-2])\d{2}[-]?\d{4}\b`), Validator: ValDKCPR, MinLength: 10, Normalization: true},
	{Type: "NO_FNR", Pattern: regexp.MustCompile(`\b(?:0[1-9]|[12]\d|3[01])(?:0[1-9]|1[0-2])\d{2}\s*\d{5}\b`), Validator: ValNone, MinLength: 11, Normalization: true},

	// Generic Entities
	{Type: "EMAIL", Pattern: regexp.MustCompile(`(?i)\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`), Validator: ValNone, MinLength: 5, Normalization: false, CaptureGroup: 0},
	{Type: "URL", Pattern: regexp.MustCompile(`(?i)https?://[^\s"<>\{\}\[\]\\]+|\bwww\.[a-zA-Z0-9\-]+\.[a-zA-Z]{2,}[^\s"<>\{\}\[\]\\]*`), Validator: ValNone, MinLength: 8, Normalization: false, CaptureGroup: 0},
	{Type: "SSN", Pattern: regexp.MustCompile(`(?i)(?:\bssn\b[:\s]*|\bsocial\s+security(?:\s+number)?(?:\s+is)?[:\s]*)\b(\d{3}-\d{2}-\d{4}|\d{9})\b`), Validator: ValNone, MinLength: 9, Normalization: true, CaptureGroup: 1},
	{Type: "SSN", Pattern: regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`), Validator: ValNone, MinLength: 11, Normalization: false, CaptureGroup: 0},
	{Type: "CREDENTIAL", Pattern: regexp.MustCompile(`(?i)\bpassword\s*[:=]\s*[^\s,]+`), Validator: ValNone, MinLength: 10, Normalization: false, CaptureGroup: 0},
	{Type: "SECRET", Pattern: regexp.MustCompile(`(?i)\b(?:secret|key|token)\s*[:=]\s*[^\s,]+`), Validator: ValNone, MinLength: 8, Normalization: false, CaptureGroup: 0},
	{Type: "PATIENT_ID", Pattern: regexp.MustCompile(`\b[A-Z]{2,3}[0-9]{6,10}\b`), Validator: ValNone, MinLength: 8, Normalization: false, CaptureGroup: 0},
	{Type: "MEDICAL_RECORD", Pattern: regexp.MustCompile(`\bMRN[- ]?[0-9]{7,10}\b`), Validator: ValNone, MinLength: 10, Normalization: false, CaptureGroup: 0},
	{Type: "PERSON", Pattern: regexp.MustCompile(`(?i)\b(?:my name is|i am|call me|this is)\s+([A-ZÀ-Ÿ][a-zà-ÿ]+(?:\s+[A-ZÀ-Ÿ][a-zà-ÿ]+){0,2})\b`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 1},

	// Legacy Scrubber Specific Entity Extraction
	{Type: "ACCOUNT_NUMBER", Pattern: regexp.MustCompile(`(?i)(compte[^n\[]*n[°o]|account[- _]?(?:number|no|nr)|num[eé]ro de compte|konto[- _]?(?:nr|nummer)|n[úu]mero de cuenta|conto corrente|n[°o]\b)\s*:?\s*([0-9]{6,20})\b`), Validator: ValNone, MinLength: 6, Normalization: false, CaptureGroup: 2},
	{Type: "MEMO_TEXT", Pattern: regexp.MustCompile(`(?i)((?:VIR(?:\s+INST)?\s+vers|WEB\s+(?:MONSIEUR|MADAME|MME|MR|MS|MRS|HERR|FRAU|SEÑOR|SEÑORA)\s+)(?:[A-ZÀ-ÖØ-Ýa-zÀ-ÖØ-öø-ÿ\-\']+\s+){1,5})([A-Za-zÀ-ÖØ-öø-ÿ][A-Za-zÀ-ÖØ-öø-ÿ0-9 ,\.\-\'!\?\/\&\*\(\)]{3,}.*)$`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 2},
	{Type: "PERSON", Pattern: regexp.MustCompile(`(?i)(?:VIR(?:\s+INST)?\s+vers\s+|(?:WEB|PISP-\w+)\s+(?:MONSIEUR|MADAME|MME|MR|MS|MRS|HERR|FRAU|SEÑOR|SEÑORA|SIGNOR|SIGNORA)?\s*|DE\s+(?:MONSIEUR|MADAME|MME|MR|MS|MRS|HERR|FRAU|SEÑOR|SEÑORA)\s+)((?:[A-ZÀ-ÖØ-Ý][A-Za-zÀ-ÖØ-öø-ÿ\-\']+(?:\s+|$)){1,4})`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 1},
	{Type: "PERSON", Pattern: regexp.MustCompile(`(?i)(SUMUP\s*\*|SUMUP  \*|SumUp\s*\*|ZETTLE_?\*|SQ\s*\*|IZETTLE\s*\*|LYRA\s*\*)([A-ZÀ-ÖØ-Ýa-zÀ-ÖØ-öø-ÿ][A-Za-zÀ-ÖØ-öø-ÿ\s\-\.]{2,35})`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 2},
	{Type: "HEALTH_ENTITY", Pattern: regexp.MustCompile(`(?i)(?:\b(ANESTHESIE|ANESTHESIA|ANäSTHESIE|ANESTESIA|CLINIQUE|CLINIC|KLINIK|CLINICA|PHARMACIE|PHARMACY|APOTHEKE|FARMACIA|CPAM|CAISSE PRIMAIRE|MUTUELLE|MUTUALITE|MUTUALIDAD|HOPITAL|HOSPITAL|KRANKENHAUS|OSPEDALE|SPITAL|MEDECIN|DOCTEUR|ARZT|MEDICO|DOTTORE|PSYCHOLOGUE|PSYCHOLOGIST|PSYCHIATRIE|PSYCHIATER|SOINS?|PFLEGE|RADIOLOGIE|RADIOLOGY|DENTISTE|DENTIST|ZAHNARZT|DENTISTA|OPTICIEN|OPTIKER|KINESITHERAPIE|PHYSIOTHERAPIE|PHYSIOTHERAPY|FISIOTERAPIA|LABORATOIRE|LABORATORIO)\b|C\.P\.A\.M\.|GROUPAMA GAN VIE|AESIO MUTUELLE|ADREA MUTUELLE)`), Validator: ValNone, MinLength: 4, Normalization: false, CaptureGroup: 0},
	// Contextual: capture the reference number that follows a health-insurance keyword (e.g. "C.P.A.M. DE L ISERE 253350004466").
	// CaptureGroup 1 tokenizes only the trailing number, leaving the keyword visible as its own HEALTH_ENTITY token.
	{Type: "HEALTH_ENTITY", Pattern: regexp.MustCompile(`(?i)(?:CPAM|C\.P\.A\.M\.|CAISSE\s+PRIMAIRE|GROUPAMA\s+GAN\s+VIE|AESIO\s+MUTUELLE|ADREA\s+MUTUELLE)\s+(?:[A-ZÀ-Ÿa-zà-ÿ\s]+\s+)?([0-9]{9,15})\b`), Validator: ValNone, MinLength: 9, Normalization: false, CaptureGroup: 1},
	{Type: "TAX_REF", Pattern: regexp.MustCompile(`(?i)\bNN[A-Z]{2}[0-9A-Z]{10,35}\b`), Validator: ValNone, MinLength: 10, Normalization: false, CaptureGroup: 0},
	{Type: "CREDITOR_REF", Pattern: regexp.MustCompile(`\b[A-Z]{2}[0-9]{2}[A-Z]{3}[0-9A-Z]{6,25}\b`), Validator: ValNone, MinLength: 10, Normalization: false, CaptureGroup: 0},
	
	// US
	{Type: "US_PASSPORT", Pattern: regexp.MustCompile(`(?i)(?:\bpassport\b|\bppt\b)[^0-9]{1,15}?([0-9]{9})\b`), Validator: ValNone, MinLength: 9, Normalization: true, CaptureGroup: 1},
	{Type: "US_DRIVERS_LICENSE_NUMBER", Pattern: regexp.MustCompile(`(?i)(?:\bdl\b|\bdriver['']?s?\s*lic(?:ense)?\b)[^A-Z0-9]{1,15}?([A-Z0-9-]{6,20})\b`), Validator: ValNone, MinLength: 6, Normalization: true, CaptureGroup: 1},

	// LATAM Core (Phase 4A)
	{Type: "BR_CPF", Pattern: regexp.MustCompile(`\b\d{3}\.?\d{3}\.?\d{3}-?\d{2}\b`), Validator: ValBRCPF, MinLength: 11, Normalization: true},
	{Type: "CL_RUT", Pattern: regexp.MustCompile(`\b(\d{1,2}(?:\.?\d{3}){2}-?[\dkK])\b`), Validator: ValCLRUT, MinLength: 8, Normalization: true},
	
	// APAC Financial (Phase 4B)
	{Type: "INDIA_AADHAAR", Pattern: regexp.MustCompile(`\b\d{12}\b`), Validator: ValIndiaAadhaar, MinLength: 12, Normalization: true},
	{Type: "SINGAPORE_ID", Pattern: regexp.MustCompile(`(?i)\b[STFGM]\d{7}[A-Z]\b`), Validator: ValSingaporeID, MinLength: 9, Normalization: true},

	// ─── P0 Tier-1 Patches ────────────────────────────────────────────────────

	// IPv6 — full 8-group form (HIPAA device identifier, GDPR, CCPA)
	{Type: "IP_ADDRESS_V6", Pattern: regexp.MustCompile(`\b[0-9A-Fa-f]{1,4}(?::[0-9A-Fa-f]{1,4}){7}\b`), Validator: ValNone, MinLength: 15, Normalization: false},
	// IPv6 — compressed (::) with ip/ipv6 keyword anchor
	{Type: "IP_ADDRESS_V6", Pattern: regexp.MustCompile(`(?i)(?:\bipv6\b|\bip\s*v6\b|\baddress\b)[:\s]+([0-9A-Fa-f]{1,4}(?::[0-9A-Fa-f]{0,4}){2,7})\b`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 1},

	// MAC address — colon or hyphen separated (HIPAA device identifier)
	{Type: "MAC_ADDRESS", Pattern: regexp.MustCompile(`\b(?:[0-9A-Fa-f]{2}[:\-]){5}[0-9A-Fa-f]{2}\b`), Validator: ValNone, MinLength: 17, Normalization: false},

	// VIN — 17-char, keyword-gated; excludes I/O/Q per ISO 3779 (HIPAA identifier)
	{Type: "VIN", Pattern: regexp.MustCompile(`(?i)(?:\bvin\b|\bvehicle\s+id(?:entification)?(?:\s+number)?)[^A-HJ-NPR-Z0-9]{0,10}([A-HJ-NPR-Z0-9]{17})\b`), Validator: ValNone, MinLength: 17, Normalization: true, CaptureGroup: 1},

	// License plate — keyword-gated (CCPA, HIPAA)
	{Type: "LICENSE_PLATE", Pattern: regexp.MustCompile(`(?i)(?:\b(?:license|licence)\s+plate\b|\bplate\s*[:#\-]|\blp\s*[:#])[^A-Z0-9]{0,8}([A-Z0-9]{2,8})\b`), Validator: ValNone, MinLength: 2, Normalization: true, CaptureGroup: 1},

	// Vehicle registration number — keyword-gated
	{Type: "VEHICLE_REG", Pattern: regexp.MustCompile(`(?i)(?:\breg(?:istration)?\s*(?:no\.?|#|number))[:\s#]*([A-Z0-9]{3,10})\b`), Validator: ValNone, MinLength: 3, Normalization: true, CaptureGroup: 1},

	// Credit card CVV/CVC — keyword-gated (PCI-DSS)
	{Type: "CARD_CVV", Pattern: regexp.MustCompile(`(?i)\b(?:cvv|cvc2?|csv|security\s+code)[:\s]*(\d{3,4})\b`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 1},

	// Credit card expiry — keyword-gated (PCI-DSS)
	{Type: "CARD_EXPIRY", Pattern: regexp.MustCompile(`(?i)(?:\bexp(?:ir(?:y|es?|ation(?:\s+date)?))?\.?\b|\bvalid(?:ity)?\s+(?:thru?|through|until)\b)[:\s]*([01]?\d[/\-][0-9]{2,4})\b`), Validator: ValNone, MinLength: 4, Normalization: false, CaptureGroup: 1},

	// Cryptocurrency wallet addresses
	// Ethereum — 0x + 40 hex chars (very low false-positive risk)
	{Type: "CRYPTO_WALLET", Pattern: regexp.MustCompile(`\b0x[0-9a-fA-F]{40}\b`), Validator: ValNone, MinLength: 42, Normalization: false},
	// Bitcoin P2PKH/P2SH — keyword-gated (base58: excludes 0, I, O, l)
	{Type: "CRYPTO_WALLET", Pattern: regexp.MustCompile(`(?i)(?:\b(?:bitcoin|btc|crypto)\b|\bwallet(?:\s+address)?\b)[:\s]*([13][a-km-zA-HJ-NP-Z1-9]{25,34}|bc1[a-z0-9]{6,87})\b`), Validator: ValNone, MinLength: 25, Normalization: false, CaptureGroup: 1},

	// SSH private key PEM header
	{Type: "SSH_KEY", Pattern: regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`), Validator: ValNone, MinLength: 27, Normalization: false},

	// IDFA / device advertising ID — keyword + UUID (COPPA, CCPA)
	{Type: "DEVICE_AD_ID", Pattern: regexp.MustCompile(`(?i)(?:\bidfa\b|\bgaid\b|\badvertising[\s_]id(?:entifier)?\b|\bdevice[\s_](?:ad[\s_])?id\b)[:\s]*([0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12})\b`), Validator: ValNone, MinLength: 36, Normalization: false, CaptureGroup: 1},

	// Date of birth — contextual, multiple formats (HIPAA, GDPR, CCPA)
	{Type: "DATE_OF_BIRTH", Pattern: regexp.MustCompile(`(?i)(?:\bdate\s+of\s+birth\b|\bdob\b|\bborn(?:\s+on)?\b)[:\s]*(\d{4}[-/\.]\d{2}[-/\.]\d{2}|\d{1,2}[-/\.]\d{1,2}[-/\.]\d{2,4}|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\.?\s+\d{1,2},?\s*\d{4})\b`), Validator: ValNone, MinLength: 6, Normalization: false, CaptureGroup: 1},

	// Hospital admission / discharge dates (HIPAA Safe Harbor)
	{Type: "CLINICAL_DATE", Pattern: regexp.MustCompile(`(?i)(?:\badmit(?:ted|tance|ting)?\b|\badmission\s+date\b|\bdischarge\s+date\b)[:\s]*(\d{4}[-/]\d{2}[-/]\d{2}|\d{1,2}[-/]\d{1,2}[-/]\d{2,4}|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\.?\s+\d{1,2},?\s*\d{4})\b`), Validator: ValNone, MinLength: 6, Normalization: false, CaptureGroup: 1},

	// Date of death (HIPAA Safe Harbor)
	{Type: "CLINICAL_DATE", Pattern: regexp.MustCompile(`(?i)(?:\bdate\s+of\s+death\b|\bdod\b|\bdeceased(?:\s+on)?\b)[:\s]*(\d{4}[-/]\d{2}[-/]\d{2}|\d{1,2}[-/]\d{1,2}[-/]\d{2,4}|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\.?\s+\d{1,2},?\s*\d{4})\b`), Validator: ValNone, MinLength: 6, Normalization: false, CaptureGroup: 1},

	// GPS coordinates with degree symbol (GDPR, CCPA precise geolocation)
	{Type: "GPS_COORD", Pattern: regexp.MustCompile(`\b\d{1,3}(?:\.\d+)?\s*°\s*[NSns][\s,;]+\d{1,3}(?:\.\d+)?\s*°\s*[EWew]\b`), Validator: ValNone, MinLength: 7, Normalization: false},
	// Decimal coordinates with keyword anchor
	{Type: "GPS_COORD", Pattern: regexp.MustCompile(`(?i)(?:\bgps\b|\bcoordinates?\b|\blat(?:itude)?\b)[:\s]*(-?\d{1,3}\.\d{4,})[\s,;]+(-?\d{1,3}\.\d{4,})\b`), Validator: ValNone, MinLength: 7, Normalization: false, CaptureGroup: 1},

	// US ZIP code — keyword-gated to avoid false positives on arbitrary 5-digit numbers
	{Type: "ZIP_CODE", Pattern: regexp.MustCompile(`(?i)(?:\bzip(?:\s+code)?\b|\bpostal\s+code\b)[:\s#]*([0-9]{5}(?:[-\s][0-9]{4})?)\b`), Validator: ValNone, MinLength: 5, Normalization: false, CaptureGroup: 1},

	// Home / residential address — street-number + name + suffix (GDPR, CCPA, HIPAA Safe Harbor)
	// Matches "742 Evergreen Terrace, Springfield" style; suffix-anchored to cut false positives.
	{Type: "HOME_ADDRESS", Pattern: regexp.MustCompile(`\b(\d{1,5}\s+[A-Z][A-Za-z]+(?:\s+[A-Za-z]+){0,4}\s+(?:Street|St|Avenue|Ave|Boulevard|Blvd|Drive|Dr|Road|Rd|Lane|Ln|Way|Court|Ct|Place|Pl|Terrace|Ter|Circle|Cir|Trail|Trl|Highway|Hwy|Parkway|Pkwy)\.?(?:\s*,\s*[A-Za-z][A-Za-z\s]{1,30})?)\b`), Validator: ValNone, MinLength: 10, Normalization: false, CaptureGroup: 1},

	// US Voter Registration ID — keyword-gated (CCPA)
	{Type: "US_VOTER_ID", Pattern: regexp.MustCompile(`(?i)(?:\bvoter\s+(?:id|#|reg(?:istration)?)\b|\bvr\s*#)[:\s#]*([A-Z]{1,4}-?\d{4,10}|\d{6,12})\b`), Validator: ValNone, MinLength: 4, Normalization: false, CaptureGroup: 1},

	// German Personalausweis — 9-char alternating letter/digit pattern: L##L##L## (GDPR)
	// Matches the fixed structure of the new German national ID number (e.g. L01X00T47).
	{Type: "DE_PERSONALAUSWEIS", Pattern: regexp.MustCompile(`\b([A-Z][0-9]{2}[A-Z][0-9]{2}[A-Z][0-9]{2})\b`), Validator: ValNone, MinLength: 9, Normalization: false, CaptureGroup: 1},

	// Professional / occupational licence — keyword-gated (GDPR, CCPA, HIPAA)
	// Catches "Bar #: CA-123456", "License No: TX-PE-9876", "Lic. 12345678", etc.
	{Type: "PROFESSIONAL_LICENSE", Pattern: regexp.MustCompile(`(?i)(?:\bbar\s*#|\blic(?:ense|ence|\.)\s*(?:#|no\.?|number)?|\bcertif(?:icate|ication)?\s*(?:#|no\.?)?|\bregistration\s*(?:#|no\.?)?|\bprofessional\s+lic(?:ense|ence)?)[:\s#]*([A-Z]{2,3}-\d{4,10}|[A-Z]{0,3}\d{5,12})\b`), Validator: ValNone, MinLength: 4, Normalization: true, CaptureGroup: 1},

	// Persistent child device identifier — child-context keyword + IDFA/GAID UUID (COPPA, GDPR Art.8)
	// Requires an explicit child-context anchor so adult device IDs are not over-flagged.
	{Type: "CHILD_DEVICE_ID", Pattern: regexp.MustCompile(`(?i)(?:\bkid(?:'?s)?\b|\bchild(?:'?s)?\b|\bjunior\b)[,\s]+(?:device|idfa|gaid|ad(?:vertising)?\s*id)[:\s]*([0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12})\b`), Validator: ValNone, MinLength: 36, Normalization: false, CaptureGroup: 1},

	// Employee / badge ID — keyword-gated (GDPR, CCPA, HIPAA)
	{Type: "EMPLOYEE_ID", Pattern: regexp.MustCompile(`(?i)(?:\bemployee(?:\s+(?:id|#|badge|no\.?))?|\bbadge(?:\s*(?:#|no\.?|number))?|\bemp(?:\s*(?:#|id|no\.?)))[:\s#]{1,3}([A-Z]{0,4}[-]?[0-9]{4,10})\b`), Validator: ValNone, MinLength: 4, Normalization: true, CaptureGroup: 1},

	// Court case / docket number — keyword-gated (CCPA, GDPR)
	{Type: "CASE_NUMBER", Pattern: regexp.MustCompile(`(?i)(?:\bcase\s+(?:no\.?|number|#)|\bdocket\s+(?:no\.?|number|#)?)[:\s#]*([A-Z0-9]{2,6}[-:\/][A-Z0-9]{2,6}(?:[-\/][A-Z0-9]{1,8})?)\b`), Validator: ValNone, MinLength: 5, Normalization: true, CaptureGroup: 1},
}

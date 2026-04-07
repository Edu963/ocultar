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

	// Financial
	{Type: "IBAN", Pattern: regexp.MustCompile(`(?i)\b[A-Z]{2}[0-9]{2}(?:[A-Z0-9]{11,30}|(?:[\s-][A-Z0-9]{4}){2,7}[\s-]?[A-Z0-9]{0,3})\b`), Validator: ValMod97, MinLength: 15, Normalization: true, CaptureGroup: 0},
	{Type: "CREDIT_CARD", Pattern: regexp.MustCompile(`\b(?:4[0-9\s-]{12,19}|5[1-5][0-9\s-]{14,19}|6(?:011|5[0-9]{2})[0-9\s-]{12,19}|3[47][0-9\s-]{13,19}|3(?:0[0-5]|[68][0-9])[0-9\s-]{11,19}|(?:2131|1800|35\d{3})[0-9\s-]{11,19})\b`), Validator: ValLuhn, MinLength: 13, Normalization: true, CaptureGroup: 0},
	{Type: "BIC", Pattern: regexp.MustCompile(`\b[A-Z]{6}[A-Z0-9]{2}(?:[A-Z0-9]{3})?\b`), Validator: ValNone, MinLength: 8, Normalization: false, CaptureGroup: 0},

	// EU/UK VAT (Generic EU VAT)
	{Type: "EU_VAT", Pattern: regexp.MustCompile(`(?i)\b(?:ATU[0-9]{8}|BE0[0-9]{9}|BG[0-9]{9,10}|CY[0-9]{8}[A-Z]|CZ[0-9]{8,10}|DE[0-9]{9}|DK[0-9]{8}|EE[0-9]{9}|EL[0-9]{9}|ES[A-Z0-9][0-9]{7}[A-Z0-9]|FI[0-9]{8}|FR[A-Z0-9]{2}[0-9]{9}|HR[0-9]{11}|HU[0-9]{8}|IE[0-9][A-Z0-9+*][0-9]{5}[A-Z]|IT[0-9]{11}|LT[0-9]{9,12}|LU[0-9]{8}|LV[0-9]{11}|MT[0-9]{8}|NL[0-9]{9}B[0-9]{2}|PL[0-9]{10}|PT[0-9]{9}|RO[0-9]{2,10}|SE[0-9]{12}|SI[0-9]{8}|SK[0-9]{10}|XI[0-9]{9})\b`), Validator: ValNone, MinLength: 6, Normalization: true},

	// France
	{Type: "FR_NIR", Pattern: regexp.MustCompile(`\b[12]\s*\d{2}\s*(?:0[1-9]|1[0-2])\s*(?:2[AB]|\d{2})\s*\d{3}\s*\d{3}\s*\d{2}\b`), Validator: ValNone, MinLength: 15, Normalization: true},
	{Type: "FRANCE_SIREN_NUMBER", Pattern: regexp.MustCompile(`\b[0-9]{3}\s*[0-9]{3}\s*[0-9]{3}\b`), Validator: ValLuhn, MinLength: 9, Normalization: true},
	{Type: "FRANCE_SIRET_NUMBER", Pattern: regexp.MustCompile(`\b[0-9]{3}\s*[0-9]{3}\s*[0-9]{3}\s*[0-9]{5}\b`), Validator: ValLuhn, MinLength: 14, Normalization: true},

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

	// Legacy Scrubber Specific Entity Extraction
	{Type: "ACCOUNT_NUMBER", Pattern: regexp.MustCompile(`(?i)(compte[^n\[]*n[°o]|account[- _]?(?:number|no|nr)|num[eé]ro de compte|konto[- _]?(?:nr|nummer)|n[úu]mero de cuenta|conto corrente|n[°o]\b)\s*:?\s*([0-9]{6,20})\b`), Validator: ValNone, MinLength: 6, Normalization: false, CaptureGroup: 2},
	{Type: "MEMO_TEXT", Pattern: regexp.MustCompile(`(?i)((?:VIR(?:\s+INST)?\s+vers|WEB\s+(?:MONSIEUR|MADAME|MME|MR|MS|MRS|HERR|FRAU|SEÑOR|SEÑORA)\s+)(?:[A-ZÀ-ÖØ-Ýa-zÀ-ÖØ-öø-ÿ\-\']+\s+){1,5})([A-Za-zÀ-ÖØ-öø-ÿ][A-Za-zÀ-ÖØ-öø-ÿ0-9 ,\.\-\'!\?\/\&\*\(\)]{3,}.*)$`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 2},
	{Type: "PERSON", Pattern: regexp.MustCompile(`(?i)(?:VIR(?:\s+INST)?\s+vers\s+|(?:WEB|PISP-\w+)\s+(?:MONSIEUR|MADAME|MME|MR|MS|MRS|HERR|FRAU|SEÑOR|SEÑORA|SIGNOR|SIGNORA)?\s*|DE\s+(?:MONSIEUR|MADAME|MME|MR|MS|MRS|HERR|FRAU|SEÑOR|SEÑORA)\s+)((?:[A-ZÀ-ÖØ-Ý][A-Za-zÀ-ÖØ-öø-ÿ\-\']+(?:\s+|$)){1,4})`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 1},
	{Type: "PERSON", Pattern: regexp.MustCompile(`(?i)(SUMUP\s*\*|SUMUP  \*|SumUp\s*\*|ZETTLE_?\*|SQ\s*\*|IZETTLE\s*\*|LYRA\s*\*)([A-ZÀ-ÖØ-Ýa-zÀ-ÖØ-öø-ÿ][A-Za-zÀ-ÖØ-öø-ÿ\s\-\.]{2,35})`), Validator: ValNone, MinLength: 3, Normalization: false, CaptureGroup: 2},
	{Type: "HEALTH_ENTITY", Pattern: regexp.MustCompile(`(?i)(?:\b(ANESTHESIE|ANESTHESIA|ANäSTHESIE|ANESTESIA|CLINIQUE|CLINIC|KLINIK|CLINICA|PHARMACIE|PHARMACY|APOTHEKE|FARMACIA|CPAM|CAISSE PRIMAIRE|MUTUELLE|MUTUALITE|MUTUALIDAD|HOPITAL|HOSPITAL|KRANKENHAUS|OSPEDALE|SPITAL|MEDECIN|DOCTEUR|ARZT|MEDICO|DOTTORE|PSYCHOLOGUE|PSYCHOLOGIST|PSYCHIATRIE|PSYCHIATER|SOINS?|PFLEGE|RADIOLOGIE|RADIOLOGY|DENTISTE|DENTIST|ZAHNARZT|DENTISTA|OPTICIEN|OPTIKER|KINESITHERAPIE|PHYSIOTHERAPIE|PHYSIOTHERAPY|FISIOTERAPIA|LABORATOIRE|LABORATORIO)\b|C\.P\.A\.M\.|GROUPAMA GAN VIE|AESIO MUTUELLE|ADREA MUTUELLE)`), Validator: ValNone, MinLength: 4, Normalization: false, CaptureGroup: 0},
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
}

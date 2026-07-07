package goprojectusat

/*
Normalize(address string) (string, error)
NormalizeCity(city string) (string, error)
NormalizeRegion(region string) (string, error)
NormalizePostalCode(postalcode string) (string, error)
NormalizeStreet(street string) (string, error)
NormalizeAdressee(addressee string) (string, error)
*/

/*
"A Project US@ standardized patient address is one that includes all required address elements and that
uses standard abbreviations as shown in [the standard
definition](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf)."
*/

/*
"If components of a patient’s address are unknown, then those fields SHOULD be left blank. If those fields
are not left blank, then UNKNOWN (spelled out, all capital letters) MUST be entered for that element in
the patient record. Patient matching algorithms SHOULD NOT match on the value UNKNOWN,
developers SHOULD flag UNKNOWN in their patient matching solution to avoid misclassification. See the
Patient Address Metadata Schema. Developers MAY indicate UNKNOWN for any component of a patient
address in accordance with the standard(s) in use (e.g., if a standard only allows numeric text in the ZIP
code field, then that field may be left blank).""

Accordingly, this library leaves unknown address fields blank and, when encountering the word "UNKNOWN", replaces it
with an empty string in the normalized value.
*/

/*
"This specification does not prescribe parsing rules, including the direction in which patient addresses are
parsed if parsing is necessary. If patient address data are captured and stored in a single string field,
where elements such as street address and city are not parsed into separate fields for the purposes of
patient matching, systems SHOULD uniformly parse data according to the following format:

Business / Firm Name - Only to be used for patient addresses containing businesses
Street Address Line  - <PRIMARY ADDRESS NUMBER><PREDIRECTIONAL><STREET NAME><SUFFIX><POSTDIRECTIONAL><SECONDARY ADDRESS IDENTIFIER><SECONDARY ADDRESS>
Last Line            - <CITY><STATE><ZIP+4>
"
*/

/*
"NON-ADDRESS INFORMATION
At times, non-address data will be captured and stored in fields intended to represent a patient’s address.
In these cases, this information SHOULD be removed. Business names are allowed as outlined in the
Standardized Patient Business Addresses section. Use of geographic features are discouraged if the
patient's record contains a street address. If the patient's record does not contain a street address, then it
is recommended that developers and health information professionals not abbreviate whatever data is
presented by the patient."
*/

/*
"LETTER CASE
Alphabetical letters SHOULD be uppercase on all lines of the address. Lowercase letters are acceptable,
provided they remain human and machine readable."

Accordingly, this library will convert all address information to uppercase in the normalized value.
*/

/*
"Diacritics SHOULD follow Appendix A for mapping guidance between letters containing diacritics and
other representations."

Note that Appendix A contains specific instructions for mapping characters tht may not be the same as other
mapping techniques for diacritics and ligatures.

This library may provide an option for specifying how to handle diacritics if alternative mappings are necessary
for a given use case.
*/

/*
"Punctuation
With the exception of the hyphen in the ZIP+4 Code and in the primary number used in the patient street
address line, punctuation SHOULD be omitted in the patient address record.
Remove special characters, multiple blanks, and punctuation as follows:
All white space characters including groups of multiple white space
characters MUST be changed to a single space, except between state
abbreviations and ZIP Codes or ZIP+4 Codes and when patients have
Canadian addresses, two spaces should between the province abbreviation
and the postal code.

* Asterisks
, Commas
. Periods
( ) Parentheses
“ “ Quotations
: Colons
; Semicolons
` Apostrophes
- Hyphens, except in the ZIP+4 Code and in the primary number used in the
patient street address line. Spaces before and after the hyphen or slashes (/)
SHOULD be removed from the address or business/firm name line. Spaces
SHOULD NOT be removed between elements, as concatenation is to be
avoided.
@ At
& Ampersand

The pound sign (#) is not considered a special character or punctuation, hence, the pound sign should not
be removed. PO Box services in some locations allow for an option to use the Post Office street address
for the address, along with the PO Box number preceded by a “#” sign. The pound sign (#) COULD be used
as a secondary unit designator if the correct designation, such as APT or STE, is not known. Unprintable
characters may be considered white space."

Although 2 spaces are allowed here, this library will normalize to 1 space after the state or province
abbreviation and a postal code in accordance with the guidance in the City Names section reqiring 1 or
more spaces.

This library may provide an option to replace "APT", "STE", and "#"" with "#"" in order to normalize data
for address matching.
*/

/*
"HYPHENATED ADDRESS RANGES
Hyphenated address ranges are prevalent in New York City (for example, 112–10 BRONX RD),
Hawaii, and areas in southern California. The hyphen in the primary range MUST NOT be removed."
*/

/*
"GRID STYLE ADDRESSES
These MAY contain significant punctuation, such as periods (for example, 39.2 RD, 39.4 RD). There
are also grid style addresses in Salt Lake City that include double directionals (for example, in 842 E
1700 S: E is a predirectional, S is a postdirectional, and 1700 is located in the street name field)."
*/

/*
"ALPHANUMERIC COMBINATIONS OF ADDRESS RANGES
Some patient addresses MAY contain a combination of alpha and numeric characters. For example,
N6W23001 BLUEMOUND RD, as found in Wisconsin and Northern Illinois. Alphanumeric address
ranges create a challenge for accurate matching."
*/

/*
"FRACTIONAL ADDRESSES
Fractional patient addresses MAY be represented as three or four character positions (for example,
123 1/2 MAIN ST). 123 1/2 takes seven character positions in the range field"
*/

/*
"Street Address Line
Each known address element MUST be segmented into individual components with one space between
each element. These components are the primary address number, predirectional, street name, suffix,
postdirectional, secondary address identifier, and secondary address. Follow guidance in the Unknown
Address section if address elements are unknown or unavailable."
*/

/*
"Primary Address Number
To standardize a patient address, the primary address number MUST be placed before the street name."
*/

/*
"Predirectional
Directional is a term used to refer to the part of the address that gives directional information for a patient
address (i.e., N, S, E, W, NE, NW, SE, SW). If a directional word is found as the first word in the street
name and there is no other directional to the left of it, then the predirectional SHOULD be abbreviated to
the appropriate one– or two–character abbreviation."

NORTH BAY STREET -> N BAY STREET
EAST END AVE -> E END AVE
*/

/*
"Numeric street names, for example, 7TH ST or SEVENTH ST, MUST be conveyed exactly as it appears
in the patient’s official identification (government issued or insurance card). Corner addresses SHOULD
be replaced by standardized street addresses if known."

This rule is impossible to implement without validating the address against the paitent's official
identification and / or trusted map data. As such, this library will leave numeric street name data unchanged
by default but may provide an option for validating an address against known map data in the future.
*/

/*
"Street Suffix Abbreviations
Street suffixes such as Boulevard and Avenue MUST be abbreviated according to the standard street
suffix abbreviations in Appendix B."
*/

/*
"Two Directionals
If two directional words appear consecutively as one or two words, before the street name or following the
street name or suffix, then the two words SHOULD become either the pre– or the post-directionals.
Exceptions are any combinations of NORTH-SOUTH or EAST–WEST as consecutive words. In these
cases, the second directional SHOULD become part of the street name and SHOULD be spelled out
completely in the street name field. Directionals SHOULD be spelled out if part of the patient street
address name.

NORTH E MAIN STREET -> NE MAIN ST
SOUTHEAST FREEWAY NORTH -> SOUTHEAST FWY N
"

Detecting when directionals are actually part of the patient street address requires reference information
such as the patient's identification of trustworthy map data. Options may be provided for how to handle
such cases.

Detecting the difference between "E Street" and "East Street" might also require validation against map data.
*/

/*
"Directional letters SHOULD NOT be combined with alphabet indicators. Directional street names
SHOULD be spelled out. Directionals SHOULD be abbreviated after the street name

COUNTY ROAD N EAST -> COUNTY ROAD NE
"
*/

/*
"Directional as Part of Street Name
If the directional word appears between the street name and the suffix, then it SHOULD appear as part of
the street name and SHOULD be spelled out in the patient record

BAY W DRIVE -> BAY WEST DRIVE
NORTH AVENUE -> NORTH AVE
"
*/

/*
"Secondary Address Unit Designators
Secondary address unit designators, such as apartment or suite, are required elements for those
patient demographic records containing secondary unit designators. Secondary address unit
designators MUST be at the end of the Patient Street Address Line. The pound sign (#) MUST
NOT be used as a secondary unit designator if the correct designation, such as APT or STE, is
known. See the Special Characters section for more information."

This rule is a refinement of the earlier rule on punctuation. As stated earlier, this library will abide by
this rule by default, but provide an option for matching purposes to always use "#" instead. Otherwise it
will be imposible to match an address where the correct designation was not known with one where the
correct designation was known.
*/

/*
"Suffixes
The suffix of the address MUST conform to the standard suffix abbreviations outlined
in Appendix B.
Two Suffixes
If an address has two consecutive words that appear in Appendix B, the second of the two words
MUST be abbreviated according to the standard suffix abbreviations and MUST be placed in the
suffix field. The first of the two words SHOULD be part of the street name, and SHOULD be
spelled out in the patient record in its entirety after the street name.

Examples:
789 MAIN AVENUE DRIVE -> 789 MAIN AVENUE DR
4513 3RD STREET CIRCLE WEST -> 4513 3RD STREET CIR W
1000 AVE E -> 1000 AVENUE E
"
*/

/*
"Highways
County, state, and local highways MUST follow the standardized format as illustrated by examples in
Appendix C. Please note that words like HIGHWAY, COUNTY, or INTERSTATE are not abbreviated
if part of the patient’s street name. More examples can be viewed in Appendix C.
"
*/

/*
"City Names
City names SHALL be spelled out in their entirety. Patient address records MUST have at least one
space between the city name, two–character state abbreviations, and ZIP+4 Code"
*/

/*
"Two Letter State and Possession Abbreviations
Names of states and U.S. possessions MUST follow the standardized abbreviations outlined in
Appendix D."
*/

/*
Military and Dept. of State Addresses
(see pkg/military/military_test.go)
*/

/*
"RURAL ROUTE ADDRESSES
The rural route number in a patient record MUST be standardized as follows:
RR ___ BOX ___

Examples:
Incorrect Form Correct Form
RURAL ROUTE 91 BOX A7 RR 91 BOX A7
RFD 82 BOX 12 RR 82 BOX 12
RD 51 # 25 RR 51 BOX 25
RFD Route 4 #87a RR 4 BOX 87A
RR 2 BOX 18 Bryan Dairy Rd RR 2 BOX 18
RR03 BOX 98D RR 3 BOX 98D

Developers:
SHOULD NOT use the words RURAL, NUMBER, NO., or the pound sign (#).
MUST NOT add a leading zero before the rural route number.
SHOULD include hyphens as part of the box number only when they are part of the address.
SHOULD change the designations RFD and RD (as a meaning for rural or rural free delivery) to RR.
SHOULD NOT allow additional designations, such as town or street names, on the patient Street Address
Line of rural route addresses."
*/

/*
GENERAL DELIVERY
Developers MUST use the words GENERAL DELIVERY, all uppercase, spelled out (no abbreviation), as
the patient street address line in the patient record if the patient has a general delivery address. Each
general delivery record SHOULD carry the –9999 add-on code. The ZIP Code or ZIP+4 Code MUST be
correctly applied for patient addresses with a general delivery. Note that General Delivery is not available
at every post office.

Example:
GEN DELIVERY, TAMPA, FL 33602 -> GENERAL DELIVERY, TAMPA FL 33602-9999
*/

/*
"POST OFFICE BOX ADDRESSES
Post Office Box addresses in a patient record MUST be standardized as follows:
Project US@ Technical Specification
23
PO BOX ______ (the actual number, numbers, or letter)

Examples:

POST OFFICE BOX 11890 -> PO BOX 11890
POST OFFICE BOX G -> PO BOX G

Developers MUST NOT add a leading zero before the post office box number.

PO Box addresses often appear with the words CALLER, FIRM CALLER, BIN, LOCKBOX, or DRAWER,
or other synonyms. When this occurs, developers MUST change these words to PO BOX in the patient
record.

PO Box services in some locations allow for an option to use the Post Office street address for the
address, along with the PO Box number preceded by a “#” sign or “UNIT” designation."
*/

/*
"PRIVATE MAILBOX ADDRESSES
Private companies offering mailbox rental services to patients are considered commercial mail receiving
agencies (CMRA). Addresses on mail received at a CMRA must adhere to specific requirements in the
use of their private mailbox number (PMB).

Patient addresses at a CMRA MUST include either the PMB identifier or the numerical identifier, followed
by the appropriate private mailbox number. Developers MUST NOT use any other identifiers.
Where the CMRA‘s physical address requires its own secondary address element, the PMB or # address
must follow the specific format rules stated below. Developers MUST NOT combine the secondary
address element of the address for the CMRA and the CMRA patient‘s private box number.
The words POST OFFICE BOX or PO BOX and the private mailbox number MUST NOT be used on the
Street Address Line. The Street Address Line is the standardized address of the private company.

PMB 234
RR 1 BOX 12
HERNDON VA 22071-2716
PMB 234
10 MAIN ST STE 11
HERNDON VA 22071-2716
123 MAIN STREET PMB 4545
HERNDON VA 22071-2716
PO BOX 159753 PMB 3571
HERNDON VA 22071-2716"
*/

/*
Puerto Rico Addresses
*/

/*
"U.S. ISLANDS AND OTHER TERRITORIES

The U.S. Virgin Islands and other territories do not use urbanizations or Spanish words. Single primary
street addresses do not have lot numbers as part of the patient addresses. These are physical identifiers.
For patient addresses to the U.S. Virgin Islands, developers MUST use VI as the correct abbreviation for
the Virgin Islands. Developers MUST NOT use USVI, VIS, VI USA, or USA VI.

Examples:
2 MOUNT ROYALE EST
CHRISTIANSTED VI 00820-4470

RR 1 BOX 6601
KINGSHILL VI 00850-9802"
*/

/*
"CANADIAN ADDRESSES
The following address format is used when the postal address delivery zone is included in the address.
Developers MUST use the standard two–character abbreviation for provinces and territories. On patient
records with addresses to Canada, developers SHOULD have two spaces between the province
abbreviation and the postal code, as shown below between “ON” and “K1A 0B1”:

Example:
1010 CLEAR STREET
OTTAWA ON K1A OB1
CANADA"
*/

/*
"OTHER INTERNATIONAL ADDRESSES
The very last (or bottom) line of an international patient address MUST contain only the COUNTRY name,
and MUST be written in full with no abbreviations and SHOULD be in capital letters. Developers MUST
NOT place the postal codes of foreign country designations on the last line of the address and MUST NOT
underline the COUNTRY name. Note that the Project US@ AHIMA Companion Guide includes guidance
and best practices on the capture and management of patient addresses in Mexico.

Examples:
HARTMANNSTRASSE 7
5300 BONN 1
GERMANY

DUBAI HOSPITAL
AL KHALEEJ STREET AL BARAHA
PO BOX 7272
DUBAI
UAE

ST VINCENTS PRIVATE HOSPITAL SYDNEY
406 VICTORIA STREET
DARLINGHURST NSW 2010
AUSTRALIA

SAINT MARIEN-HOSPITAL BERLIN
GALLWITZALLEE 123/143
12249 BERLIN
GERMANY"
*/

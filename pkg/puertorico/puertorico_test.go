package puertorico_test

/*
"PUERTO RICO ADDRESSES

Puerto Rico‘s common addressing consists of various formats, such as:

Urbanization
House Number and Street Name
City, State, and ZIP+4

URB
LAS GLADIOLAS
150 CALLE A
SAN JUAN PR 00926-3232

House Number and Street Name
City, State, and ZIP+4 Code

1234 CALLE AURORA
MAYAGUEZ PR 00680-1233

Exceptions

Some areas in Puerto Rico do not have street names or repetitive house numbers. The
urbanization name SHOULD substitute as the street name.

House number and Urbanization Name
City, State, and ZIP+4

1234 URB LOS OLMOS
PONCE PR 00731-1235

There are also public housing projects (residenciales) without street names or repetitive
apartment numbers. In these cases the apartment number SHOULD be the primary number and
the name of the public housing project SHOULD become the street name.

Apartment Number and Residential Name
City, State, and ZIP+4

23 RES LLORENS TORRES
SAN JUAN PR 00924-1234

Certain condominiums are not located on a named street or have an assigned number to the
building. The name of the condominium SHOULD be substituted for the street name.

Residential Name
Building No. and Apt. No.
City, State, and ZIP+4

The word CALLE MAY be placed before the street name and number. CALLE means STREET in
Spanish, and placing the word CALLE prior to other address components is proper use based on
Spanish composition. In addition to the word CALLE, the word AVENIDA or its abbreviation AVE
MAY also appear in this position.

Apartment Buildings and Condominiums

There are two basic address formats for apartment buildings and condominiums. Developers
MUST follow abbreviation guidance outlined in the Secondary Address Unit Designators section
for patient addresses located within apartment buildings and condominiums.

Buildings with a physical street address

Building Name
Street Number, Street Name, Apartment Number
City, State, and ZIP+4

COND ASHFORD PALACE
1234 AVE ASHFORD APT 1A
SAN JUAN PR 00907-1234

Buildings without a physical address

Certain condominiums are located on an unnamed street and may not have an assigned number.
The name of the condominium SHOULD substitute as the street name and the number 1
SHOULD be used when no building number exists.

Bldg Number, Bldg Name, and Apt Number
City, State, and ZIP+4

1 COND MIRAFLOR APT 104
SAN JUAN PR 000907-1335

Where there are multiple buildings (or towers) with the same name, the building number
SHOULD become the primary number.

COND VERDE APT 1120       ->  1 COND VERDE APT 1120
VISTA SUITES III APT 104  ->  3 VISTA SUITES APT 104

Patient Street Address Line

The components of the patient Street Address Line are the urbanization (when required), primary
address number and street name, secondary address identifier, and secondary address range.

Urbanization Name
Secondary Address Identifier and Number
Primary Address Number and Street Name

URB HIGHLAND GDNS
COND LAS AMAPOLAS APT 103
123 CALLE MAIN

In Puerto Rico, some apartment buildings do not have a street address. In this situation, the
building name SHOULD be part of the primary address identifier. If directionals are present in an
address, they are part of the street name. Developers MUST NOT translate directionals.

1510 CALLE 3 NO (NO = Northwest)
1620 CALLE 17 SO (SO = Southwest)

Street Names and Prefixes

Developers MUST NOT abbreviate street names.
Spanish street names generally have the suffix element preceding the root street name, making it
a prefix.

CALLE AVENIDA, PASEO, PLAZA, PASAJE, CARR, PARQUE, VEREDA, VISTA, VIA, CALLE JON, PATIO, BLVD, CAMINO, CAMINITO, CALETA, MARGINAL

585 AVE FD ROOSEVELT
105 CAMINO AMAZONA
1025 PARQUE DEL REY
1212 VIA ANGÉLICA

Developers MUST NOT translate CALLE to the suffix ST. This translation will lead to additional
errors when matching patient records.
Note that patient addresses that will also be used for billing purposes or other mailing SHOULD
always include CALLE, AVENIDA, etc.

Numbered Streets

Numbered streets MUST always contain the word CALLE. This avoids misinterpretation between
numbered streets and house numbers in patient addresses.

Incorrect Form Correct Form
CALLE 1 A17    ->  A17 CALLE 1
CALLE 191 B113 ->  13 CALLE 191

House numbers may have fractional or alphabetic modifiers. Developers MUST place the house number
before the street name. When placing alphanumeric house numbers prior to the street name, developers
MUST NOT use hyphens to separate the letter from the number.

Incorrect Form Correct Form
CALLE 125 C-19      ->  C19 CALLE 125
A-17 CALLE AMAPOLA  ->  A17 CALLE AMAPOLA
B-17A CALLE 1       ->  B17A CALLE 1

Due to the amount of numbers within a block and a house number in Puerto Rico addresses, many
identifiers are commonly used to separate address elements, including BLOQUE, NUM, NO, CASA,
LOTE, or a # sign. These identifiers MUST NOT be included in patient addresses.
Hyphens in the address range are sometimes necessary. When addresses contain block numbers and
house numbers, developers MUST use a hyphen to separate the block number from the house number.
When addresses contain up to three–digit numeric block numbers, developers MUST include a hyphen.

Incorrect Form Correct Form
CALLE 19 BLQ 199 Casa 31    ->  199-31 CALLE 19
CALLE 117 Bloque 23 Núm.18  ->  23-18 CALLE 117

Urbanizations

Urbanization denotes an area, sector, or development within a geographic area. In addition to
being a descriptive word, it precedes the name of the area. This URB descriptor, commonly used
in urban areas of Puerto Rico, is an important part of the addressing format, as it describes the
location of a given street.

Urbanizations MUST be abbreviated to URB followed by the urbanization name. Urbanizations are not
repeated within five–digit zones.

Incorrect Form Correct Form
URBANIZATION GOLDEN GATE  ->  URB GOLDEN GATE

In Puerto Rico, identical street names and address number ranges can be found within the same ZIP
Code. In these cases, the urbanization name is the only element that correctly identifies the location of a
particular address.

URB ROYAL OAKS
123 CALLE 1
BAYAMON PR 00961-0123

URB HERMOSILLO
123 CALLE 1
BAYAMON PR 00961-1212

Exceptions

Certain urbanizations are known as extensiones, mansiones, repartos, villas, parques, and jardines.
When these names are present in a patient address, MUST NOT place the abbreviation URB prior to the
name of the urbanization. Some addresses in Puerto Rico urbanizations do not have a street name,
where the urbanization MUST become the street name.

Incorrect Form Correct Form
A17 URB JARDINES FAGOTA    ->  A17 JARD FAGOTA
PONCE PR 00731                 PONCE PR 00731

The following urbanization names stand alone and MUST NOT require the use of the abbreviation
URB. Abbreviations containing the letter S in parentheses at the end of the abbreviation allows for
the plural representation of the word in an abbreviated form.

Urbanization Abbreviation
Altura(s) ALT(S)
Barriada BDA
Barrio BO
Bosque BOSQUE
Brisa(s) BRISA(S)
Ciudad CIUDAD
Colina(s) COLINA(S)
Chalets CHALETS
Comunidad COMUNIDAD
Estancias EST
Extensión EXT
Hacienda HACIENDA
Jardines JARD
Industrial IND
Loma(s) LOMA(S)
Mansiones MANS
Parque PARQ
Parcela(s) PARCELA(S)
Paseo PASEO
Pradera PRADERA
Portal PORTAL
Portales PORTALES
Quintas QUINTAS
Residencial RES
Reparto REPTO
Riberas RIBERAS
Sector SECT
Terraza TERR
Valle VALLE
Villa(s) VILLA(S)
Vista(s) VISTA(S)

Examples:
URB EXT VISTA BELLA   ->  EXT VISTA BELLA
URB ALTS DE CANÁ      ->  ALTS DE CANA

Post Office Box

Developers MUST capture or transform Post Office Box addresses as PO BOX in the patient record.
Developers MUST NOT use Spanish words to represent PO BOX.

XYZ COMPANY    ->  XYZ COMPANY
APARTADO 2018      PO BOX 2018

ABC COMPANY    ->  ABC COMPANY
GPO BOX 1118       PO BOX 1118

In certain areas, the postal station name appears in a patient’s address. The postal station name is not
needed because the ZIP Code identifies the station. However, when the station name is present, it
SHOULD be placed above the delivery line.

PO BOX 1190                 OLD SAN JUAN STA
OLD SAN JUAN STA        ->  PO BOX 1190
SAN JUAN PR 00902-1190      SAN JUAN PR 00902-1190

Rural Routes

A rural route address in the patient record MUST be standardized as follows:

RR___ BOX___
Developers MUST NOT use the words RURAL, RUTA RURAL, BUZON, or BZN. The
designations RFD, RD, and RT (meaning rural route) MUST be changed to RR and developers
MUST have a space between RR and the route number and BOX and the box number.
Developers MUST NOT add a leading zero before the rural route number.

RR03 BOX 9800            -> RR 3 BOX 9800
RFD ROUTE 4 BZN 1725     -> RR 4 BOX 1725
RUTA RURAL 3 BUZON 12000 -> RR 3 BOX 12000
RFD 1 Bzn 17-A           -> RR 1 BOX 17A

There MUST NOT be additional designations, such as sector names, on the Street Address Line
of rural route addresses.

Names of sectors used together with route and box numbers can lead to increased matching
errors. Health IT developers MUST eliminate this information in Puerto Rico addresses.

RR 2 BOX 1980
SECTOR EL BRINCO -> RR 2 BOX 1980

RR 3 BOX 3415
BARRIO VISTA ALEGRE -> RR 3 BOX 3415

Highway Contract Routes

Highway contract route addresses MUST be standardized as HC____BOX____. It is basically the
same format utilized for rural routes. Likewise, Health IT developers MUST NOT include leading
zeros before the route number.

Ruta Estrella 1 Buzón 18 -> HC 1 BOX 18
HC 03 Bzn 1050           -> HC 3 BOX 1050

As with rural route addresses, developers MUST NOT include any additional designations, such
as names of sectors in the patient address line of highway contract addresses.

Last Line
Patient addresses SHOULD include the last line, which MUST include the city, state and ZIP
Code, if known. Certain areas of the San Juan metropolitan area are identified by residents with
names such as Condado, Barrio Obrero, and Rio Piedras. Developers MUST NOT use these
names to represent the city of San Juan. These are not valid last line entries. Developers MUST
include SAN JUAN as the only valid city name for patient addresses within San Juan"
*/

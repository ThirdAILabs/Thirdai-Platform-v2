const a = [
    {
      'taggedTokens': [
        ['Agent', 'O'], [':', 'O'], ['Thank', 'O'], ['you', 'O'], ['for', 'O'], ['calling', 'O'], ['TechGuard', 'O'],
        ['Support', 'O'], ['.', 'O'], ['My', 'O'], ['name', 'O'], ['is', 'O'], ['Alex', 'O'], [',', 'O'], ['how', 'O'],
        ['can', 'O'], ['I', 'O'], ['help', 'O'], ['you', 'O'], ['today', 'O'], ['?', 'O']
      ],
      'sourceObject': 'customer_chat_20231105_0.txt',
      'groups': ['Safe']
    },
    {
      'taggedTokens': [
        ['Customer', 'O'], [':', 'O'], ['Hi', 'O'], ['there', 'O'], ['.', 'O'], ['My', 'O'], ['name', 'O'], ['is', 'O'],
        ['Robert', 'NAME'], ['Chen', 'NAME'], ['and', 'O'], ["I'm", 'O'], ['having', 'O'], ['trouble', 'O'],
        ['accessing', 'O'], ['my', 'O'], ['account', 'O'], ['.', 'O'], ["I've", 'O'], ['been', 'O'], ['trying', 'O'],
        ['since', 'O'], ['yesterday', 'O'], ['.', 'O']
      ],
      'sourceObject': 'customer_chat_20231105_1.txt',
      'groups': ['Review']
    },
    {
      'taggedTokens': [
        ['Agent', 'O'], [':', 'O'], ["I'm", 'O'], ['sorry', 'O'], ['to', 'O'], ['hear', 'O'], ['that', 'O'], [',', 'O'],
        ['Mr.', 'O'], ['Chen', 'NAME'], ['.', 'O'], ["I'd", 'O'], ['be', 'O'], ['happy', 'O'], ['to', 'O'], ['help', 'O'],
        ['you', 'O'], ['regain', 'O'], ['access', 'O'], ['.', 'O'], ['Could', 'O'], ['you', 'O'], ['please', 'O'],
        ['verify', 'O'], ['your', 'O'], ['account', 'O'], ['with', 'O'], ['your', 'O'], ['email', 'O'], ['address', 'O'],
        ['?', 'O']
      ],
      'sourceObject': 'customer_chat_20231105_2.txt',
      'groups': ['Review']
    }, 
    {
      'taggedTokens': [
        ['Customer', 'O'], [':', 'O'], ['Sure', 'O'], [',', 'O'], ["it's", 'O'], ['robert.chen1982', 'EMAIL'], ['@', 'EMAIL'], ['gmail.com', 'EMAIL'], ['.', 'O']
      ], 'sourceObject': 'customer_chat_20231105_3.txt', 'groups': ['Review']
    }, { 'taggedTokens': [['Agent', 'O'], [':', 'O'], ['Thank', 'O'], ['you', 'O'], ['.', 'O'], ['And', 'O'], ['for', 'O'], ['additional', 'O'], ['verification', 'O'], [',', 'O'], ['could', 'O'], ['I', 'O'], ['have', 'O'], ['the', 'O'], ['last', 'O'], ['four', 'O'], ['digits', 'O'], ['of', 'O'], ['the', 'O'], ['phone', 'O'], ['number', 'O'], ['associated', 'O'], ['with', 'O'], ['the', 'O'], ['account', 'O'], ['?', 'O']], 'sourceObject': 'customer_chat_20231105_4.txt', 'groups': ['Safe'] }, { 'taggedTokens': [['Customer', 'O'], [':', 'O'], ['Yes', 'O'], [',', 'O'], ["it's", 'O'], ['5784', 'PHONE'], ['.', 'O']], 'sourceObject': 'customer_chat_20231105_5.txt', 'groups': ['Review'] }, { 'taggedTokens': [['Agent', 'O'], [':', 'O'], ['Perfect', 'O'], ['.', 'O'], ['I', 'O'], ['can', 'O'], ['see', 'O'], ['your', 'O'], ['account', 'O'], ['here', 'O'], ['.', 'O'], ['It', 'O'], ['looks', 'O'], ['like', 'O'], ['there', 'O'], ['were', 'O'], ['multiple', 'O'], ['failed', 'O'], ['login', 'O'], ['attempts', 'O'], ['from', 'O'], ['an', 'O'], ['unfamiliar', 'O'], ['IP', 'O'], ['address', 'O'], [',', 'O'], ['so', 'O'], ['our', 'O'], ['security', 'O'], ['system', 'O'], ['temporarily', 'O'], ['locked', 'O'], ['your', 'O'], ['account', 'O'], ['.', 'O'], ['Can', 'O'], ['you', 'O'], ['confirm', 'O'], ['your', 'O'], ['current', 'O'], ['address', 'O'], ['is', 'O'], ['still', 'O'], ['728', 'ADDRESS'], ['Maple', 'ADDRESS'], ['Street', 'ADDRESS'], [',', 'ADDRESS'], ['Apartment', 'ADDRESS'], ['4B', 'ADDRESS'], [',', 'ADDRESS'], ['San', 'ADDRESS'], ['Francisco', 'ADDRESS'], [',', 'ADDRESS'], ['CA', 'ADDRESS'], ['94107', 'ADDRESS'], ['?', 'O']], 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive'] }, { 'taggedTokens': [['Customer', 'O'], [':', 'O'], ['Yes', 'O'], [',', 'O'], ["that's", 'O'], ['correct', 'O'], ['.', 'O']], 'sourceObject': 'customer_chat_20231105_7.txt', 'groups': ['Safe'] }, { 'taggedTokens': [['Agent', 'O'], [':', 'O'], ['Great', 'O'], ['.', 'O'], ["I've", 'O'], ['reset', 'O'], ['your', 'O'], ['account', 'O'], ['access', 'O'], ['.', 'O'], ['You', 'O'], ['should', 'O'], ['receive', 'O'], ['a', 'O'], ['verification', 'O'], ['code', 'O'], ['at', 'O'], ['your', 'O'], ['email', 'O'], ['address', 'O'], ['shortly', 'O'], ['.', 'O'], ['Once', 'O'], ['you', 'O'], ['enter', 'O'], ['that', 'O'], ['code', 'O'], [',', 'O'], ["you'll", 'O'], ['be', 'O'], ['prompted', 'O'], ['to', 'O'], ['create', 'O'], ['a', 'O'], ['new', 'O'], ['password', 'O'], ['.', 'O'], ['Is', 'O'], ['there', 'O'], ['anything', 'O'], ['else', 'O'], ['I', 'O'], ['can', 'O'], ['help', 'O'], ['with', 'O'], ['today', 'O'], ['?', 'O']], 'sourceObject': 'customer_chat_20231105_8.txt', 'groups': ['Safe'] }, { 'taggedTokens': [['Customer', 'O'], [':', 'O'], ['Actually', 'O'], [',', 'O'], ['yes', 'O'], ['.', 'O'], ['I', 'O'], ['recently', 'O'], ['got', 'O'], ['a', 'O'], ['new', 'O'], ['credit', 'O'], ['card', 'O'], ['and', 'O'], ['need', 'O'], ['to', 'O'], ['update', 'O'], ['my', 'O'], ['billing', 'O'], ['information', 'O'], ['.', 'O'], ['The', 'O'], ['new', 'O'], ['card', 'O'], ['number', 'O'], ['is', 'O'], ['4832', 'CREDIT_CARD'], ['5691', 'CREDIT_CARD'], ['2748', 'CREDIT_CARD'], ['1035', 'CREDIT_CARD'], ['with', 'O'], ['expiration', 'O'], ['date', 'O'], ['09', 'EXPIRATION_DATE'], ['/', 'EXPIRATION_DATE'], ['27', 'EXPIRATION_DATE'], ['and', 'O'], ['security', 'O'], ['code', 'O'], ['382', 'CVV'], ['.', 'O']], 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive'] }, { 'taggedTokens': [['Customer', 'O'], [':', 'O'], ['That', 'O'], ['makes', 'O'], ['sense', 'O'], ['.', 'O'], ["I'll", 'O'], ['do', 'O'], ['that', 'O'], ['instead', 'O'], ['.', 'O'], ['My', 'O'], ['social', 'O'], ['security', 'O'], ['number', 'O'], ['is', 'O'], ['532', 'SSN'], ['-', 'SSN'], ['48', 'SSN'], ['-', 'SSN'], ['1095', 'SSN'], ['if', 'O'], ['you', 'O'], ['need', 'O'], ['that', 'O'], ['for', 'O'], ['verification', 'O'], ['.', 'O']], 'sourceObject': 'customer_chat_20231105_10.txt', 'groups': ['Sensitive'] },
    {
      'taggedTokens': [
        ['Hi', 'O'], ['my', 'O'], ['name', 'O'], ['is', 'O'], ['John', 'NAME'], ['Smith', 'NAME'], ['and', 'O'],
        ['my', 'O'], ['phone', 'O'], ['number', 'O'], ['is', 'O'], ['555-123-4567', 'PHONE']
      ],
      'sourceObject': 'customer_chat_20231105.txt',
      'groups': ['Sensitive'],
    },
    {
      'taggedTokens': [
        ['Please', 'O'], ['update', 'O'], ['my', 'O'], ['address', 'O'], ['to', 'O'],
        ['123', 'ADDRESS'], ['Main', 'ADDRESS'], ['Street', 'ADDRESS'], ['Apt', 'ADDRESS'], ['4B', 'ADDRESS']
      ],
      'sourceObject': 'customer_chat_20231106.txt',
      'groups': ['Safe'],
    },
    {
      'taggedTokens': [
        ['My', 'O'], ['SSN', 'O'], ['is', 'O'], ['123-45-6789', 'SSN'], ['.', 'O'],
        ['Date', 'O'], ['of', 'O'], ['birth', 'O'], ['is', 'O'], ['01/15/1980', 'DOB']
      ],
      'sourceObject': 'customer_chat_20231107.txt',
      'groups': ['Sensitive', 'Sensitive'],
    },
    {
      'taggedTokens': [
        ['I', 'O'], ['need', 'O'], ['help', 'O'], ['with', 'O'], ['my', 'O'], ['insurance', 'O'], ['claim', 'O'],
        ['987654321', 'CLAIM_ID']
      ],
      'sourceObject': 'customer_chat_20231108.txt',
      'groups': ['Sensitive'],
    },
    {
      'taggedTokens': [
        ['You', 'O'], ['can', 'O'], ['email', 'O'], ['me', 'O'], ['at', 'O'],
        ['john.smith', 'EMAIL'], ['@', 'EMAIL'], ['healthcare.com', 'EMAIL']
      ],
      'sourceObject': 'customer_chat_20231109.txt',
      'groups': ['Safe'],
    },
    {
      'taggedTokens': [
        ['My', 'O'], ['policy', 'O'], ['number', 'O'], ['is', 'O'],
        ['POL987654321', 'POLICY_NUMBER'], ['issued', 'O'], ['by', 'O'], ['BlueCross', 'O']
      ],
      'sourceObject': 'customer_chat_20231110.txt',
      'groups': ['Sensitive'],
    },
  ];
  
  
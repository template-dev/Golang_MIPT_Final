const GATEWAY_URL = "http://localhost:8080";
const AUTH_URL = "http://localhost:8081";

function onOpen() {
  SpreadsheetApp.getUi()
    .createMenu("CashApp")
    .addItem("Register", "registerUser")
    .addItem("Login", "login")
    .addSeparator()
    .addItem("Push Budgets", "pushBudgets")
    .addItem("Push Transactions", "addTransactions")
    .addItem("Bulk Transactions", "bulkTransactions")
    .addSeparator()
    .addItem("Load Report", "loadReport")
    .addToUi();
}

function getAuthSheet_() {
  return SpreadsheetApp.getActive().getSheetByName("Auth");
}

function getToken_() {
  const sh = getAuthSheet_();
  return sh.getRange("C2").getValue(); // access token
}

function setToken_(access) {
  const sh = getAuthSheet_();
  if (access) sh.getRange("C2").setValue(access);
}

function getCreds_() {
  const sh = getAuthSheet_();
  const email = sh.getRange("A2").getValue();
  const password = sh.getRange("B2").getValue();
  return { email, password };
}

function registerUser() {
  const { email, password } = getCreds_();
  const resp = UrlFetchApp.fetch(AUTH_URL + "/auth/register", {
    method: "post",
    contentType: "application/json",
    payload: JSON.stringify({ email, password }),
    muteHttpExceptions: true
  });
  if (resp.getResponseCode() !== 201 && resp.getResponseCode() !== 200) {
    throw new Error(resp.getContentText());
  }
  SpreadsheetApp.getUi().alert("Registered");
}

function login() {
  const { email, password } = getCreds_();
  const resp = UrlFetchApp.fetch(AUTH_URL + "/auth/login", {
    method: "post",
    contentType: "application/json",
    payload: JSON.stringify({ email, password }),
    muteHttpExceptions: true
  });
  if (resp.getResponseCode() !== 200) {
    throw new Error(resp.getContentText());
  }
  const data = JSON.parse(resp.getContentText());
  setToken_(data.access_token);
  SpreadsheetApp.getUi().alert("Logged in");
}

function authHeaders_() {
  const access = getToken_();
  if (!access) throw new Error("JWT token missing. Run login()");
  return { Authorization: "Bearer " + access };
}

function pushBudgets() {
  const ss = SpreadsheetApp.getActive();
  const sh = ss.getSheetByName("Budgets");
  const rows = sh.getDataRange().getValues();
  const headers = authHeaders_();

  for (let i = 1; i < rows.length; i++) {
    const [category, limit] = rows[i];
    if (!category || !limit) continue;

    const resp = UrlFetchApp.fetch(GATEWAY_URL + "/api/budgets", {
      method: "post",
      contentType: "application/json",
      headers,
      payload: JSON.stringify({ category, limit }),
      muteHttpExceptions: true
    });

    const statusCell = sh.getRange(i + 1, 3);
    statusCell.setValue(resp.getResponseCode() === 201 ? "OK" : resp.getContentText());
  }
}

function addTransactions() {
  const ss = SpreadsheetApp.getActive();
  const sh = ss.getSheetByName("Transactions");
  const rows = sh.getDataRange().getValues();
  const headers = authHeaders_();

  for (let i = 1; i < rows.length; i++) {
    const [amount, category, description, date] = rows[i];
    if (!amount || !category || !date) continue;

    const resp = UrlFetchApp.fetch(GATEWAY_URL + "/api/transactions", {
      method: "post",
      contentType: "application/json",
      headers,
      payload: JSON.stringify({
        amount,
        category,
        description: description || "",
        date: date
      }),
      muteHttpExceptions: true
    });

    const statusCell = sh.getRange(i + 1, 5);
    statusCell.setValue(resp.getResponseCode() === 201 ? "OK" : resp.getContentText());
  }
}

function bulkTransactions() {
  const ss = SpreadsheetApp.getActive();
  const sh = ss.getSheetByName("Transactions");
  const rows = sh.getDataRange().getValues();
  const headers = authHeaders_();

  const items = [];
  const idxMap = [];

  for (let i = 1; i < rows.length; i++) {
    const [amount, category, description, date] = rows[i];
    if (!amount || !category || !date) continue;
    idxMap.push(i);
    items.push({ amount, category, description: description || "", date });
  }

  if (items.length === 0) {
    SpreadsheetApp.getUi().alert("No items to import");
    return;
  }

  const resp = UrlFetchApp.fetch(GATEWAY_URL + "/api/transactions/bulk?workers=4", {
    method: "post",
    contentType: "application/json",
    headers,
    payload: JSON.stringify(items),
    muteHttpExceptions: true
  });

  if (resp.getResponseCode() !== 200) {
    throw new Error(resp.getContentText());
  }

  const out = JSON.parse(resp.getContentText());
  const errByIndex = {};
  (out.errors || []).forEach(e => errByIndex[e.index] = e.error);

  for (let j = 0; j < idxMap.length; j++) {
    const rowIdx = idxMap[j];
    const statusCell = sh.getRange(rowIdx + 1, 5);
    statusCell.setValue(errByIndex[j] ? errByIndex[j] : "OK");
  }

  SpreadsheetApp.getUi().alert(`Bulk done. accepted=${out.accepted}, rejected=${out.rejected}`);
}

function loadReport() {
  const ss = SpreadsheetApp.getActive();
  const headers = authHeaders_();

  const from = "2025-12-01";
  const to = "2025-12-31";

  const resp = UrlFetchApp.fetch(`${GATEWAY_URL}/api/reports/summary?from=${from}&to=${to}`, {
    headers,
    muteHttpExceptions: true
  });

  if (resp.getResponseCode() !== 200) {
    throw new Error(resp.getContentText());
  }

  const data = JSON.parse(resp.getContentText());
  const sh = ss.getSheetByName("Report");
  sh.clear();
  sh.getRange(1, 1, 1, 2).setValues([["Category", "Total"]]);

  const keys = Object.keys(data).sort();
  let row = 2;
  keys.forEach(k => {
    sh.getRange(row, 1).setValue(k);
    sh.getRange(row, 2).setValue(data[k]);
    row++;
  });
}

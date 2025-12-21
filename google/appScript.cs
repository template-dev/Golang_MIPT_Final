const GATEWAY_URL = "http://localhost:8080";
const AUTH_URL = "http://localhost:8081";

function login() {
  const sheet = SpreadsheetApp.getActive().getSheetByName("Auth");
  const email = sheet.getRange("A2").getValue();
  const password = sheet.getRange("B2").getValue();

  const resp = UrlFetchApp.fetch(AUTH_URL + "/auth/login", {
    method: "post",
    contentType: "application/json",
    payload: JSON.stringify({
      email: email,
      password: password
    }),
    muteHttpExceptions: true
  });

  if (resp.getResponseCode() !== 200) {
    throw new Error(resp.getContentText());
  }

  const token = JSON.parse(resp.getContentText()).access_token;
  sheet.getRange("C2").setValue(token);
}

function addTransactions() {
  const ss = SpreadsheetApp.getActive();
  const authSheet = ss.getSheetByName("Auth");
  const token = authSheet.getRange("C2").getValue();

  if (!token) {
    throw new Error("JWT token missing. Run login()");
  }

  const sheet = ss.getSheetByName("Transactions");
  const rows = sheet.getDataRange().getValues();

  for (let i = 1; i < rows.length; i++) {
    const [amount, category, description, date] = rows[i];
    if (!amount || !category) continue;

    const resp = UrlFetchApp.fetch(GATEWAY_URL + "/api/transactions", {
      method: "post",
      contentType: "application/json",
      headers: {
        Authorization: "Bearer " + token
      },
      payload: JSON.stringify({
        amount: amount,
        category: category,
        description: description || "",
        date: date
      }),
      muteHttpExceptions: true
    });

    if (resp.getResponseCode() === 201) {
      sheet.getRange(i + 1, 5).setValue("OK");
    } else {
      sheet.getRange(i + 1, 5).setValue(resp.getContentText());
    }
  }
}

function loadReport() {
  const ss = SpreadsheetApp.getActive();
  const authSheet = ss.getSheetByName("Auth");
  const token = authSheet.getRange("C2").getValue();

  const from = "2025-12-01";
  const to = "2025-12-31";

  const resp = UrlFetchApp.fetch(
    `${GATEWAY_URL}/api/reports/summary?from=${from}&to=${to}`,
    {
      headers: {
        Authorization: "Bearer " + token
      },
      muteHttpExceptions: true
    }
  );

  if (resp.getResponseCode() !== 200) {
    throw new Error(resp.getContentText());
  }

  const data = JSON.parse(resp.getContentText());
  const sheet = ss.getSheetByName("Report");
  sheet.clear();
  sheet.getRange(1, 1, 1, 2).setValues([["Category", "Total"]]);

  let row = 2;
  for (const k in data) {
    sheet.getRange(row, 1).setValue(k);
    sheet.getRange(row, 2).setValue(data[k]);
    row++;
  }
}

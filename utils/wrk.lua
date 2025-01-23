wrk.method = "POST"
wrk.path = "/cart/checkout"
wrk.body = "email=test@example.com&street_address=123 Main St&zip_code=98101&city=Seattle&state=WA&country=USA&credit_card_number=4111111111111111&credit_card_expiration_month=12&credit_card_expiration_year=2025&credit_card_cvv=123"
wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"

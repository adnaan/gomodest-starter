```bash
stripe products create --name="Free" --description="Free plan"
stripe prices create \
  -d product=prod_IZQ0fAMigCUFX1 \
  -d unit_amount=0 \
  -d currency=eur \
  -d "recurring[interval]"=month
stripe products create --name="Pro" --description="Pro plan"
stripe prices create \
  -d product=prod_IZQ2RckEQabB3y \
  -d unit_amount=10 \
  -d currency=eur \
  -d "recurring[interval]"=month
```
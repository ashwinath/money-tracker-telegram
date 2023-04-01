# Money Tracker Telegram

## Introduction

To account for everyday spending, I have been writing into my saved messages and parsing them every month. This is extremely tedious. The aim here is to write a small parser to understand how I spend

## Telegram API Definition

### Help

Typing `help` or any unknown command will return the help interface.

### Types

Type | Explanation
-----|------------
reim | Amount to be reimbursed, usually paying first using CC and friends paying back later.
shared reim | Amount to be reimbursed, usually paying first using CC and taking from shared account.
special shared reim | Amount to be reimbursed, usually paying first using CC and taking from shared account, not counting into regular spend.
shared | Amount that is shared but other party had paid first
special shared | Amount that is shared but other party had paid first and it's a one off thing
own | Regular type of spending for ownself.

### Classification

Classification could be any string field. This is for your own note taking and not used in a special way.

Class | Explanation
------|------------
meal | Amount spent on meals
housing | Amount spent on housing needs

### Adding a transaction

User: `ADD <TYPE> <CLASSIFICATION> <PRICE (no $ sign)>`
Service returns: `Created Transaction ID: <ID>, Transaction: <transaction>`

### Deleting a transaction

User: `DEL <ID>`
Service returns: `Deleted Transaction ID: <ID>, Transaction: <transaction>`

## Web scraping (internal network only)

To be continued.

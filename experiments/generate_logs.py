import random

first_names = ["John", "Sarah", "Michael", "Emma", "David", "Jessica"]
last_names = ["Doe", "Smith", "Johnson", "Williams", "Brown", "Jones"]
domains = ["example.com", "test.org", "company.net", "corp.mail"]
phones = ["+1-555-0198", "408-555-0178", "+44-20-7946-0958", "06 12 34 56 78"]

with open("massive_test.csv", "w") as f:
    f.write("timestamp,ip,action,user_name,email,phone,notes\n")
    for i in range(100000):  # 100k lines
        fname = random.choice(first_names)
        lname = random.choice(last_names)
        email = f"{fname.lower()}.{lname.lower()}@{random.choice(domains)}"
        phone = random.choice(phones)
        f.write(
            f"2026-02-21T10:00:00Z,192.168.1.{random.randint(1,254)},LOGIN_ATTEMPT,{fname} {lname},{email},{phone},User {fname} logged in from {email} calling {phone}.\n"
        )
print("Generated massive_test.csv (100,000 lines)")

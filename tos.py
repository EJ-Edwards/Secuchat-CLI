import os
import textwrap

TERMS = """
Secuchat-CLI — Usage Agreement
===============================

This self-hosted communication tool is for authorized red team operations only.

1. AUTHORIZED PERSONNEL ONLY
   - Access restricted to approved red team members
   - Must have current authorization for ongoing engagements
   - Unauthorized personnel are prohibited from using this system

2. ENGAGEMENT SCOPE COMPLIANCE
   - All communications must relate to authorized penetration testing activities
   - Stay within the defined scope of current engagements
   - Immediately cease communications if engagement scope changes or ends

3. OPERATIONAL SECURITY (OPSEC)
   - Maintain strict OPSEC protocols in all communications
   - Use appropriate code names and operational terminology
   - No real client names, IP addresses, or sensitive identifiers in plain text
   - Follow organization's communication security guidelines

4. PROFESSIONAL CONDUCT
   - Maintain professional standards at all times
   - No inappropriate, offensive, or unprofessional content
   - Respect all team members and operational requirements
   - Report security incidents or policy violations immediately

5. DATA HANDLING
   - No transmission of actual client data or credentials
   - Use this channel for coordination and tactical communication only
   - Sensitive findings should be documented through secure reporting channels
   - Follow organization's data classification and handling policies

6. SYSTEM SECURITY
   - Report any technical issues or security concerns immediately
   - Do not attempt to bypass or modify security controls
   - Use strong authentication credentials
   - Log out properly when sessions end

7. INCIDENT RESPONSE
   - Report compromise or unauthorized access immediately
   - Cease operations if legal or safety concerns arise
   - Follow organization's incident response procedures
   - Maintain operational logs as required

By accepting, you confirm you are authorized to participate in current 
red team operations and will use this tool in accordance with organizational 
policies and engagement parameters.
"""

def clear_screen():
    os.system("cls" if os.name == "nt" else "clear")

def accept_terms():
    clear_screen()
    print(textwrap.dedent(TERMS))
    while True:
        c = input("Accept terms? (y/n): ").strip().lower()
        if c in ("y", "yes"): return True
        if c in ("n", "no"): return False
        print("Enter y or n.")

def main():
    if accept_terms():
        print("✅ Terms accepted.")
        exit(0)
    else:
        print("❌ Terms not accepted.")
        exit(1)

if __name__ == "__main__":
    main()
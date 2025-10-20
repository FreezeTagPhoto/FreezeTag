
import xml.etree.ElementTree as ET

tree = ET.parse('coverage-full.xml')

coverageElem = tree.getroot()
if coverageElem == None:
    print("error no coverage: coverage: 0%")
    exit(1)
coverage = coverageElem.get('line-rate')
if coverage == None:
    print("error no line-rate: coverage: 0%")
    exit(1)

print(f"coverage: {float(coverage) * 100}%")
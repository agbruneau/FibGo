import os
import re
from pathlib import Path
from collections import defaultdict

def count_words_in_file(file_path):
    """Compte les mots dans un fichier Markdown en ignorant la syntaxe."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Supprimer les blocs de code
        content = re.sub(r'```[\s\S]*?```', '', content)
        content = re.sub(r'`[^`]+`', '', content)
        
        # Supprimer les liens markdown [texte](url)
        content = re.sub(r'\[([^\]]+)\]\([^\)]+\)', r'\1', content)
        
        # Supprimer les images ![alt](url)
        content = re.sub(r'!\[([^\]]*)\]\([^\)]+\)', '', content)
        
        # Supprimer les titres markdown (conservant le texte)
        content = re.sub(r'^#+\s+', '', content, flags=re.MULTILINE)
        
        # Supprimer les caractères de formatage markdown
        content = re.sub(r'\*\*([^\*]+)\*\*', r'\1', content)  # bold
        content = re.sub(r'\*([^\*]+)\*', r'\1', content)      # italic
        content = re.sub(r'__([^_]+)__', r'\1', content)        # bold
        content = re.sub(r'_([^_]+)_', r'\1', content)          # italic
        
        # Supprimer les séparateurs horizontaux
        content = re.sub(r'^---+$', '', content, flags=re.MULTILINE)
        content = re.sub(r'^===+$', '', content, flags=re.MULTILINE)
        
        # Supprimer les tableaux markdown (lignes de séparation)
        content = re.sub(r'^\|[\s\-\|:]+$', '', content, flags=re.MULTILINE)
        
        # Supprimer les caractères spéciaux markdown
        content = re.sub(r'[><]', '', content)  # citations
        
        # Compter les mots (séparés par des espaces ou sauts de ligne)
        # Extraction des mots en utilisant \b pour les limites de mots
        words = re.findall(r'\b\w+\b', content)
        return len(words)
    except Exception as e:
        print(f"Erreur lors de la lecture de {file_path}: {e}")
        return 0

def count_words_in_volume(volume_path):
    """Compte les mots dans tous les fichiers .md d'un volume."""
    total_words = 0
    file_count = 0
    files = []
    
    for root, dirs, filenames in os.walk(volume_path):
        for filename in filenames:
            if filename.endswith('.md'):
                file_path = os.path.join(root, filename)
                words = count_words_in_file(file_path)
                total_words += words
                file_count += 1
                relative_path = os.path.relpath(file_path, volume_path)
                files.append((relative_path, words))
    
    return total_words, file_count, sorted(files)

# Parcourir les volumes
volumes = [
    "Volume_I_Fondations_Entreprise_Agentique",
    "Volume_II_Infrastructure_Agentique",
    "Volume_III_Apache_Kafka_Guide_Architecte",
    "Volume_IV_Apache_Iceberg_Lakehouse",
    "Volume_V_Developpeur_Renaissance"
]

results = {}

print("Analyse en cours...\n")

for volume in volumes:
    if os.path.exists(volume):
        words, count, files = count_words_in_volume(volume)
        results[volume] = {
            'total_words': words,
            'file_count': count,
            'files': files
        }
        print(f"[OK] {volume}: {count} fichiers, {words:,} mots")
    else:
        print(f"[ERREUR] Volume introuvable: {volume}")

# Afficher les résultats détaillés
print("\n" + "="*80)
print("COMPTAGE DES MOTS PAR VOLUME")
print("="*80 + "\n")

total_all_volumes = 0

for volume, data in results.items():
    volume_name = volume.replace("Volume_", "V").replace("_", " ")
    print(f"{volume_name}:")
    print(f"  Nombre de fichiers: {data['file_count']}")
    print(f"  Nombre total de mots: {data['total_words']:,}")
    print()
    total_all_volumes += data['total_words']

print("="*80)
print(f"TOTAL TOUS VOLUMES: {total_all_volumes:,} mots")
print("="*80)

# Tableau récapitulatif
print("\n" + "="*80)
print("RÉCAPITULATIF")
print("="*80)
print(f"{'Volume':<50} {'Fichiers':<10} {'Mots':>15}")
print("-"*80)
for volume, data in results.items():
    volume_name = volume.replace("Volume_", "V").replace("_", " ")
    print(f"{volume_name:<50} {data['file_count']:<10} {data['total_words']:>15,}")
print("-"*80)
print(f"{'TOTAL':<50} {sum(r['file_count'] for r in results.values()):<10} {total_all_volumes:>15,}")
print("="*80)

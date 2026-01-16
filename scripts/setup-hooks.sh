#!/bin/bash

# Script para configurar Git Hooks
# Executa testes automaticamente antes de cada commit

echo "🔧 Configurando Git Hooks..."

# Verificar se estamos em um repositório Git
if [ ! -d ".git" ]; then
    echo "❌ Erro: Este não é um repositório Git!"
    exit 1
fi

# Criar diretório de hooks se não existir
mkdir -p .git/hooks

# Copiar pre-commit hook
echo "📋 Instalando pre-commit hook..."
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

echo ""
echo "🧪 Running tests before commit..."
echo ""

# Executar testes
make test

# Verificar se os testes passaram
if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Tests failed! Commit aborted."
    echo "💡 Fix the tests before committing."
    echo "💡 To bypass (use with caution): git commit --no-verify"
    echo ""
    exit 1
fi

echo ""
echo "✅ All tests passed! Proceeding with commit..."
echo ""
exit 0
EOF

# Tornar o hook executável
chmod +x .git/hooks/pre-commit

echo ""
echo "✅ Git Hooks configurados com sucesso!"
echo ""
echo "📝 O que foi configurado:"
echo "   • pre-commit: Executa 'make test' antes de cada commit"
echo ""
echo "💡 Dica: Para fazer commit sem executar testes (emergências):"
echo "   git commit --no-verify -m \"mensagem\""
echo ""

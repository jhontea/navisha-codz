terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
  }
  backend "s3" {
    bucket = "coding-challenge-terraform-state"
    key    = "prod/terraform.tfstate"
    region = "ap-southeast-1"
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  name = "coding-challenge-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway     = true
  single_nat_gateway     = false
  enable_dns_hostnames   = true
  enable_dns_support     = true
  enable_vpn_gateway     = false

  tags = {
    Environment = "production"
    Project     = "coding-challenge"
  }
}

# EKS Cluster
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "19.15.0"

  cluster_name    = "coding-challenge-eks"
  cluster_version = "1.28"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access = true

  eks_managed_node_groups = {
    api = {
      desired_size = 3
      min_size     = 2
      max_size     = 6
      instance_types = ["t3.medium"]
      labels = {
        role = "api"
      }
    }
    worker = {
      desired_size = 2
      min_size     = 1
      max_size     = 10
      instance_types = ["c5.xlarge"]
      labels = {
        role = "worker"
      }
    }
    sandbox = {
      desired_size = 2
      min_size     = 1
      max_size     = 20
      instance_types = ["c5.2xlarge"]
      labels = {
        role = "sandbox"
      }
    }
  }

  tags = {
    Environment = "production"
    Project     = "coding-challenge"
  }
}

# RDS PostgreSQL
resource "aws_db_instance" "primary" {
  identifier = "coding-challenge-db"

  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.r6g.large"

  db_name  = "coding_challenge"
  username = var.db_username
  password = var.db_password

  allocated_storage     = 100
  max_allocated_storage = 500
  storage_type          = "gp3"
  storage_encrypted     = true

  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.rds.name

  backup_retention_period = 30
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"

  multi_az               = true
  deletion_protection    = true
  skip_final_snapshot    = false
  final_snapshot_identifier = "coding-challenge-db-final"

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  tags = {
    Name        = "coding-challenge-db"
    Environment = "production"
  }
}

# ElastiCache Redis
resource "aws_elasticache_replication_group" "redis" {
  replication_group_id = "coding-challenge-redis"
  description         = "Redis cluster for coding challenge"

  node_type            = "cache.r6g.large"
  num_cache_clusters   = 3
  port                 = 6379

  parameter_group_name = "default.redis7"
  engine               = "redis"
  engine_version       = "7.1"

  subnet_group_name          = aws_elasticache_subnet_group.redis.name
  security_group_ids         = [aws_security_group.redis.id]
  automatic_failover_enabled = true
  multi_az_enabled           = true

  at_rest_encryption_enabled  = true
  transit_encryption_enabled  = true

  tags = {
    Name        = "coding-challenge-redis"
    Environment = "production"
  }
}

# Amazon MQ RabbitMQ
resource "aws_mq_broker" "rabbitmq" {
  broker_name = "coding-challenge-mq"

  engine_type        = "RabbitMQ"
  engine_version     = "3.12"
  host_instance_type = "mq.m5.large"

  auto_minor_version_upgrade = true
  deployment_mode           = "CLUSTER_MULTI_AZ"

  subnet_ids = module.vpc.private_subnets
  security_groups = [aws_security_group.rabbitmq.id]

  user {
    username = var.rabbitmq_username
    password = var.rabbitmq_password
  }

  logs {
    general = true
    audit   = true
  }

  tags = {
    Name        = "coding-challenge-mq"
    Environment = "production"
  }
}

# ECR Repositories
resource "aws_ecr_repository" "services" {
  for_each = toset(["auth-service", "problem-service", "execution-service",
                     "execution-worker", "leaderboard-service", "hint-service",
                     "websocket-service", "api-gateway"])

  name                 = "coding-challenge/${each.key}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

# Security Groups
resource "aws_security_group" "rds" {
  name        = "coding-challenge-rds"
  description = "RDS PostgreSQL security group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port = 5432
    to_port   = 5432
    protocol  = "tcp"
    tags = { Name = "PostgreSQL" }
  }
}

resource "aws_security_group" "redis" {
  name        = "coding-challenge-redis"
  description = "Redis security group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port = 6379
    to_port   = 6379
    protocol  = "tcp"
  }
}

resource "aws_security_group" "rabbitmq" {
  name        = "coding-challenge-rabbitmq"
  description = "RabbitMQ security group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port = 5671
    to_port   = 5672
    protocol  = "tcp"
  }
  ingress {
    from_port = 15671
    to_port   = 15672
    protocol  = "tcp"
  }
}
